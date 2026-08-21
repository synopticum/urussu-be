package service

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"urussu-be/internal/auth"
	"urussu-be/internal/domain"
)

const (
	// pipelineTimeout bounds the whole async pipeline so goroutines cannot
	// leak forever.
	pipelineTimeout = 10 * time.Second
	// statusUpdateTimeout bounds the final status write; it runs on a fresh
	// context because the pipeline context may already be cancelled.
	statusUpdateTimeout = 5 * time.Second

	// Emulated durations of the external calls, until a real mailer and a
	// real audit log replace them.
	emailDelay = 2 * time.Second
	auditDelay = 2 * time.Second
)

// commentIDKey carries the comment ID through the detached pipeline context
// so a future tracing layer can pick it up without structural changes.
type commentIDKey struct{}

// CommentsRepository is the storage contract required by CommentsService.
// Defined on the consumer side so the service can be tested with a stub.
type CommentsRepository interface {
	List(ctx context.Context, entityID string) ([]domain.Comment, error)
	Create(ctx context.Context, userID, entityID string, entityType domain.CommentEntityType, body string) (domain.Comment, error)
	GetStatus(ctx context.Context, id string) (domain.CommentStatus, error)
	UpdateStatus(ctx context.Context, id string, status domain.CommentStatus) error
}

// CommentsService implements the comments use-cases. Creating a comment
// additionally launches a detached async pipeline (email notification +
// audit log entry, both emulated) that flips the comment's status column
// when done.
type CommentsService struct {
	repo CommentsRepository
	log  *slog.Logger

	// pipelineParent is the app lifecycle context: pipeline goroutines must
	// not derive from a request context (grpc-gateway cancels it when the
	// handler returns), but must still observe shutdown (SIGINT/SIGTERM).
	pipelineParent context.Context

	// wg tracks in-flight pipelines so Shutdown can wait for their final
	// status UPDATEs before the pgx pool is closed.
	wg sync.WaitGroup
}

func NewCommentsService(repo CommentsRepository, log *slog.Logger, pipelineParent context.Context) *CommentsService {
	return &CommentsService{repo: repo, log: log, pipelineParent: pipelineParent}
}

func (s *CommentsService) ListComments(ctx context.Context, entityID string) ([]domain.Comment, error) {
	return s.repo.List(ctx, entityID)
}

// Timeline example for the async detached pipeline below
// — happy path —
//
// t=0: CreateComment: WithTimeout(parent, 10s) → goroutine starts
// t=2s: email + audit timers fire → tasks return nil
// t≈2s: status UPDATE (fresh 5s ctx) → wg.Done() → cancel() stops the 10s timer

// — or the unhappy path —

// t=10s: deadline fires → ctx.Done() closes → tasks return DeadlineExceeded
// t≈10s: status UPDATE "failed" on fresh 5s ctx → wg.Done() → cancel()

// — or shutdown at t=3s —

// parent canceled → child canceled instantly (timer stopped early)
// → tasks interrupted → status UPDATE "failed" on fresh ctx → Shutdown's
//   wg.Wait() unblocks → pool closes

func (s *CommentsService) CreateComment(ctx context.Context, userID, entityID string, entityType domain.CommentEntityType, body string) (domain.Comment, error) {
	comment, err := s.repo.Create(ctx, userID, entityID, entityType, body)
	if err != nil {
		return domain.Comment{}, err
	}

	// The values (user ID, comment ID) are copied into the detached context
	// so pipeline logs carry them and a future tracing layer produces child
	// spans sharing one trace_id with the handler.
	pipelineCtx, cancel := context.WithTimeout(s.pipelineParent, pipelineTimeout)
	pipelineCtx = auth.ContextWithUser(pipelineCtx, userID, auth.RoleFrom(ctx))
	pipelineCtx = context.WithValue(pipelineCtx, commentIDKey{}, comment.ID)

	s.wg.Add(1)
	go func() {
		defer cancel()
		s.runPipeline(pipelineCtx, comment.ID, userID)
	}()

	return comment, nil
}

func (s *CommentsService) GetCommentStatus(ctx context.Context, id string) (domain.CommentStatus, error) {
	return s.repo.GetStatus(ctx, id)
}

// Shutdown waits for in-flight pipelines to finish their final status
// UPDATEs. It must be called before the pgx pool is closed.
func (s *CommentsService) Shutdown(ctx context.Context) error {
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// runPipeline runs the post-processing tasks concurrently and records the
// outcome in the comment's status column: completed when every task
// succeeded, failed otherwise.
func (s *CommentsService) runPipeline(ctx context.Context, commentID, userID string) {
	defer s.wg.Done()

	attrs := []any{slog.String("user_id", userID), slog.String("comment_id", commentID)}

	var tasks sync.WaitGroup
	tasks.Add(2)
	var emailErr, auditErr error
	go func() {
		defer tasks.Done()
		emailErr = s.sendEmailNotification(ctx, attrs)
	}()
	go func() {
		defer tasks.Done()
		auditErr = s.writeAuditLogEntry(ctx, attrs)
	}()
	tasks.Wait()

	status := domain.CommentStatusCompleted
	if emailErr != nil || auditErr != nil {
		status = domain.CommentStatusFailed
	}

	// The pipeline context may already be done (timeout or shutdown), but
	// the status must still reach the DB, so the write gets a fresh context.
	updateCtx, cancel := context.WithTimeout(context.Background(), statusUpdateTimeout)
	defer cancel()
	if err := s.repo.UpdateStatus(updateCtx, commentID, status); err != nil {
		s.log.ErrorContext(ctx, "failed to update comment status", append(attrs, slog.Any("error", err))...)
		return
	}

	s.log.InfoContext(ctx, "comment pipeline finished", append(attrs, slog.String("status", string(status)))...)
}

// sendEmailNotification emulates emailing the comment author: a timer plus
// a log entry, no real mailer. The delay is interruptible so shutdown and
// the pipeline timeout are observed immediately.
func (s *CommentsService) sendEmailNotification(ctx context.Context, attrs []any) error {
	s.log.InfoContext(ctx, "sending comment email notification", attrs...)

	select {
	case <-time.After(emailDelay):
		s.log.InfoContext(ctx, "comment email notification sent", attrs...)
		return nil
	case <-ctx.Done():
		s.log.InfoContext(ctx, "comment email notification interrupted", append(attrs, slog.Any("error", ctx.Err()))...)
		return ctx.Err()
	}
}

// writeAuditLogEntry emulates writing an audit log entry: a timer plus a
// log entry, no audit table. A real audit log (and tracing) will replace it.
func (s *CommentsService) writeAuditLogEntry(ctx context.Context, attrs []any) error {
	s.log.InfoContext(ctx, "writing comment audit log entry", attrs...)

	select {
	case <-time.After(auditDelay):
		s.log.InfoContext(ctx, "comment audit log entry written", attrs...)
		return nil
	case <-ctx.Done():
		s.log.InfoContext(ctx, "comment audit log write interrupted", append(attrs, slog.Any("error", ctx.Err()))...)
		return ctx.Err()
	}
}
