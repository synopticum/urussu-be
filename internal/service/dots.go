// Package service contains the use-case layer between transport handlers
// and repositories.
package service

import (
	"context"

	"urussu-be/internal/domain"
)

// DotsRepository is the storage contract required by DotsService.
// Defined on the consumer side so the service can be tested with a stub.
type DotsRepository interface {
	List(ctx context.Context, limit int32) ([]domain.Dot, error)
}

// DotsService implements the dots use-cases.
type DotsService struct {
	repo DotsRepository
}

func NewDotsService(repo DotsRepository) *DotsService {
	return &DotsService{repo: repo}
}

func (s *DotsService) ListDots(ctx context.Context, limit int32) ([]domain.Dot, error) {
	return s.repo.List(ctx, limit)
}
