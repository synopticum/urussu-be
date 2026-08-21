// Package postgres implements the PostgreSQL-backed repositories.
package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
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

func (r *DotsRepository) GetByID(ctx context.Context, id string) (domain.Dot, error) {
	var d domain.Dot

	err := r.pool.QueryRow(ctx,
		`SELECT d.id::text, d.title, d.description, l.name, d.coordinates
		 FROM dots d
		 JOIN layers l ON l.id = d.layer
		 WHERE d.id = $1`, id).
		Scan(&d.ID, &d.Title, &d.Description, &d.Layer, &d.Coordinates)

	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Dot{}, domain.ErrNotFound
	}

	if err != nil {
		return domain.Dot{}, fmt.Errorf("query dot %s: %w", id, err)
	}

	return d, nil
}

func (r *DotsRepository) List(ctx context.Context, limit int32, layer string) ([]domain.Dot, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT d.id::text, d.title, d.description, l.name, d.coordinates
		 FROM dots d
		 JOIN layers l ON l.id = d.layer
		 WHERE ($2::text = '' OR l.name = $2)
		 LIMIT $1`, limit, layer)
	if err != nil {
		return nil, fmt.Errorf("query dots: %w", err)
	}
	defer rows.Close()

	var dots []domain.Dot
	for rows.Next() {
		var d domain.Dot
		if err := rows.Scan(&d.ID, &d.Title, &d.Description, &d.Layer, &d.Coordinates); err != nil {
			return nil, fmt.Errorf("scan dot: %w", err)
		}
		dots = append(dots, d)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate dots: %w", err)
	}

	return dots, nil
}
