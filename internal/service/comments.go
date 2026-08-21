package service

import (
	"context"

	"urussu-be/internal/domain"
)

// CommentsRepository is the storage contract required by CommentsService.
// Defined on the consumer side so the service can be tested with a stub.
type CommentsRepository interface {
	List(ctx context.Context, entityID string) ([]domain.Comment, error)
	Create(ctx context.Context, userID, entityID string, entityType domain.CommentEntityType, body string) (domain.Comment, error)
}

// CommentsService implements the comments use-cases.
type CommentsService struct {
	repo CommentsRepository
}

func NewCommentsService(repo CommentsRepository) *CommentsService {
	return &CommentsService{repo: repo}
}

func (s *CommentsService) ListComments(ctx context.Context, entityID string) ([]domain.Comment, error) {
	return s.repo.List(ctx, entityID)
}

func (s *CommentsService) CreateComment(ctx context.Context, userID, entityID string, entityType domain.CommentEntityType, body string) (domain.Comment, error) {
	return s.repo.Create(ctx, userID, entityID, entityType, body)
}
