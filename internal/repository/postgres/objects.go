package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"urussu-be/internal/domain"
)

// ObjectsRepository reads objects from PostgreSQL.
type ObjectsRepository struct {
	pool *pgxpool.Pool
}

func NewObjectsRepository(pool *pgxpool.Pool) *ObjectsRepository {
	return &ObjectsRepository{pool: pool}
}

func (r *ObjectsRepository) GetByID(ctx context.Context, id string) (domain.Object, error) {
	var o domain.Object
	var imagesJSON []byte

	err := r.pool.QueryRow(ctx,
		`SELECT o.id::text, o.house, o.street, o.description, o.radius, o.coordinates,
		        (SELECT jsonb_agg(jsonb_build_object('id', i.id::text, 'year', i.year) ORDER BY i.year)
		         FROM images i
		         WHERE i.entity_type = 'object' AND i.entity_id = o.id::text)
		 FROM objects o
		 WHERE o.id = $1`, id).
		Scan(&o.ID, &o.House, &o.Street, &o.Description, &o.Radius, &o.Coordinates, &imagesJSON)

	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Object{}, domain.ErrNotFound
	}

	if err != nil {
		return domain.Object{}, fmt.Errorf("query object %s: %w", id, err)
	}

	o.Images, err = decodeImages(imagesJSON)
	if err != nil {
		return domain.Object{}, fmt.Errorf("query object %s: %w", id, err)
	}

	return o, nil
}

func (r *ObjectsRepository) List(ctx context.Context, limit int32) ([]domain.Object, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT o.id::text, o.house, o.street, o.description, o.radius, o.coordinates,
		        (SELECT jsonb_agg(jsonb_build_object('id', i.id::text, 'year', i.year) ORDER BY i.year)
		         FROM images i
		         WHERE i.entity_type = 'object' AND i.entity_id = o.id::text)
		 FROM objects o
		 LIMIT $1`, limit)
	if err != nil {
		return nil, fmt.Errorf("query objects: %w", err)
	}
	defer rows.Close()

	var objects []domain.Object
	for rows.Next() {
		var o domain.Object
		var imagesJSON []byte
		if err := rows.Scan(&o.ID, &o.House, &o.Street, &o.Description, &o.Radius, &o.Coordinates, &imagesJSON); err != nil {
			return nil, fmt.Errorf("scan object: %w", err)
		}
		o.Images, err = decodeImages(imagesJSON)
		if err != nil {
			return nil, fmt.Errorf("scan object: %w", err)
		}
		objects = append(objects, o)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate objects: %w", err)
	}

	return objects, nil
}
