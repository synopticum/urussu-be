package service

import (
	"context"

	"urussu-be/internal/domain"
)

// ObjectsRepository is the storage contract required by ObjectsService.
// Defined on the consumer side so the service can be tested with a stub.
type ObjectsRepository interface {
	GetByID(ctx context.Context, id string) (domain.Object, error)
	List(ctx context.Context, limit int32) ([]domain.Object, error)
}

// ObjectsService implements the objects use-cases.
type ObjectsService struct {
	repo ObjectsRepository
}

func NewObjectsService(repo ObjectsRepository) *ObjectsService {
	return &ObjectsService{repo: repo}
}

func (s *ObjectsService) GetObject(ctx context.Context, id string) (domain.Object, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *ObjectsService) ListObjects(ctx context.Context, limit int32) ([]domain.Object, error) {
	if limit <= 0 {
		limit = 3000
	}

	return s.repo.List(ctx, limit)
}
