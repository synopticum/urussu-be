package grpc

import (
	"context"
	"log/slog"
	"unicode/utf8"

	urussuv1 "urussu-be/gen/urussu/v1"
	"urussu-be/internal/auth"
	"urussu-be/internal/domain"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// maxCommentBodyLen mirrors the VARCHAR(240) limit of the comments.body
// column so oversized bodies are rejected before hitting the database.
const maxCommentBodyLen = 240

// CommentsService is the use-case contract required by CommentsHandler.
// Defined on the consumer side so the handler can be tested with a stub.
type CommentsService interface {
	ListComments(ctx context.Context, entityID string) ([]domain.Comment, error)
	CreateComment(ctx context.Context, userID, entityID string, entityType domain.CommentEntityType, body string) (domain.Comment, error)
}

// CommentsHandler implements urussuv1.CommentsServiceServer.
type CommentsHandler struct {
	urussuv1.UnimplementedCommentsServiceServer

	svc CommentsService
	log *slog.Logger
}

func NewCommentsHandler(svc CommentsService, log *slog.Logger) *CommentsHandler {
	return &CommentsHandler{svc: svc, log: log}
}

func (h *CommentsHandler) ListComments(ctx context.Context, req *urussuv1.ListCommentsRequest) (*urussuv1.ListCommentsResponse, error) {
	if req.GetEntityId() == "" {
		return nil, status.Error(codes.InvalidArgument, "entity_id is required")
	}

	comments, err := h.svc.ListComments(ctx, req.GetEntityId())
	if err != nil {
		h.log.ErrorContext(ctx, "failed to list comments", slog.String("entity_id", req.GetEntityId()), slog.Any("error", err))
		return nil, status.Error(codes.Internal, "failed to list comments")
	}

	resp := &urussuv1.ListCommentsResponse{Comments: make([]*urussuv1.Comment, 0, len(comments))}
	for _, c := range comments {
		resp.Comments = append(resp.Comments, commentToProto(c))
	}

	return resp, nil
}

func (h *CommentsHandler) CreateComment(ctx context.Context, req *urussuv1.CreateCommentRequest) (*urussuv1.CreateCommentResponse, error) {
	if req.GetEntityId() == "" {
		return nil, status.Error(codes.InvalidArgument, "entity_id is required")
	}

	var entityType domain.CommentEntityType
	switch req.GetEntityType() {
	case urussuv1.CommentEntityType_COMMENT_ENTITY_TYPE_DOT:
		entityType = domain.CommentEntityDot
	case urussuv1.CommentEntityType_COMMENT_ENTITY_TYPE_OBJECT:
		entityType = domain.CommentEntityObject
	case urussuv1.CommentEntityType_COMMENT_ENTITY_TYPE_PATH:
		entityType = domain.CommentEntityPath
	default:
		return nil, status.Error(codes.InvalidArgument, "entity_type must be one of DOT, OBJECT or PATH")
	}

	// Rune count matches how PostgreSQL measures VARCHAR(240).
	if req.GetBody() == "" {
		return nil, status.Error(codes.InvalidArgument, "body is required")
	}
	if utf8.RuneCountInString(req.GetBody()) > maxCommentBodyLen {
		return nil, status.Errorf(codes.InvalidArgument, "body exceeds %d characters", maxCommentBodyLen)
	}

	// The author always comes from the authenticated session, never from the
	// request; the JWT middleware guarantees it for this endpoint.
	userID := auth.UserIDFrom(ctx)
	if userID == "" {
		return nil, status.Error(codes.Unauthenticated, "authentication required")
	}

	comment, err := h.svc.CreateComment(ctx, userID, req.GetEntityId(), entityType, req.GetBody())
	if err != nil {
		h.log.ErrorContext(ctx, "failed to create comment", slog.String("entity_id", req.GetEntityId()), slog.Any("error", err))
		return nil, status.Error(codes.Internal, "failed to create comment")
	}

	return &urussuv1.CreateCommentResponse{Comment: commentToProto(comment)}, nil
}

func commentToProto(c domain.Comment) *urussuv1.Comment {
	return &urussuv1.Comment{
		Id:         c.ID,
		Name:       c.Name,
		Body:       c.Body,
		CreatedAt:  timestamppb.New(c.CreatedAt),
		ModifiedAt: timestamppb.New(c.ModifiedAt),
	}
}
