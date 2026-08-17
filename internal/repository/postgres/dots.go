// Package postgres implements the PostgreSQL-backed repositories.
package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"urussu-be/internal/domain"
)

// DotsRepository reads dots from PostgreSQL.
type DotsRepository struct {
	pool *pgxpool.Pool
}

func NewDotsRepository(pool *pgxpool.Pool) *DotsRepository {
	return &DotsRepository{pool: pool}
}

func (r *DotsRepository) List(ctx context.Context, limit int32) ([]domain.Dot, error) {
	if limit <= 0 {
		limit = 10
	}

	rows, err := r.pool.Query(ctx,
		`SELECT id::text, title, description, layer::text FROM dots LIMIT $1`, limit)
	if err != nil {
		return nil, fmt.Errorf("query dots: %w", err)
	}
	defer rows.Close()

	var dots []domain.Dot
	for rows.Next() {
		var d domain.Dot
		if err := rows.Scan(&d.ID, &d.Title, &d.ShortDescription, &d.Layer); err != nil {
			return nil, fmt.Errorf("scan dot: %w", err)
		}
		dots = append(dots, d)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate dots: %w", err)
	}

	return dots, nil
}
