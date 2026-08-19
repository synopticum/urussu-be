package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"urussu-be/internal/domain"
)

// PathsRepository reads paths from PostgreSQL.
type PathsRepository struct {
	pool *pgxpool.Pool
}

func NewPathsRepository(pool *pgxpool.Pool) *PathsRepository {
	return &PathsRepository{pool: pool}
}

func (r *PathsRepository) GetByID(ctx context.Context, id string) (domain.Path, error) {
	var p domain.Path

	err := r.pool.QueryRow(ctx,
		`SELECT p.id::text, p.title, p.description, p.coordinates
		 FROM paths p
		 WHERE p.id = $1`, id).
		Scan(&p.ID, &p.Title, &p.Description, &p.Coordinates)

	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Path{}, domain.ErrNotFound
	}

	if err != nil {
		return domain.Path{}, fmt.Errorf("query path %s: %w", id, err)
	}

	return p, nil
}

func (r *PathsRepository) List(ctx context.Context, limit int32) ([]domain.Path, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT p.id::text, p.title, p.description, p.coordinates
		 FROM paths p
		 LIMIT $1`, limit)
	if err != nil {
		return nil, fmt.Errorf("query paths: %w", err)
	}
	defer rows.Close()

	var paths []domain.Path
	for rows.Next() {
		var p domain.Path
		if err := rows.Scan(&p.ID, &p.Title, &p.Description, &p.Coordinates); err != nil {
			return nil, fmt.Errorf("scan path: %w", err)
		}
		paths = append(paths, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate paths: %w", err)
	}

	return paths, nil
}
