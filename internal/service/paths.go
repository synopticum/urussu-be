package service

import (
	"context"

	"urussu-be/internal/domain"
)

// PathsRepository is the storage contract required by PathsService.
// Defined on the consumer side so the service can be tested with a stub.
type PathsRepository interface {
	GetByID(ctx context.Context, id string) (domain.Path, error)
	List(ctx context.Context, limit int32) ([]domain.Path, error)
}

// PathsService implements the paths use-cases.
type PathsService struct {
	repo PathsRepository
}

func NewPathsService(repo PathsRepository) *PathsService {
	return &PathsService{repo: repo}
}

func (s *PathsService) GetPath(ctx context.Context, id string) (domain.Path, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *PathsService) ListPaths(ctx context.Context, limit int32) ([]domain.Path, error) {
	if limit <= 0 {
		limit = 1000
	}

	return s.repo.List(ctx, limit)
}
