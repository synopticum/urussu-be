package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"urussu-be/internal/domain"
)

// CommentsRepository reads comments from PostgreSQL.
type CommentsRepository struct {
	pool *pgxpool.Pool
}

func NewCommentsRepository(pool *pgxpool.Pool) *CommentsRepository {
	return &CommentsRepository{pool: pool}
}

// List returns the comments of the given entity. A comment belongs to
// exactly one entity, so matching against any FK column is enough.
func (r *CommentsRepository) List(ctx context.Context, entityID string) ([]domain.Comment, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT c.id::text, u.first_name || ' ' || u.last_name, c.body, c.created_at, c.modified_at
		 FROM comments c
		 JOIN users u ON u.id = c.user_id
		 WHERE c.dot_id = $1 OR c.object_id = $1 OR c.path_id = $1
		 ORDER BY c.created_at`, entityID)
	if err != nil {
		return nil, fmt.Errorf("query comments for entity %s: %w", entityID, err)
	}
	defer rows.Close()

	var comments []domain.Comment
	for rows.Next() {
		var c domain.Comment
		if err := rows.Scan(&c.ID, &c.Name, &c.Body, &c.CreatedAt, &c.ModifiedAt); err != nil {
			return nil, fmt.Errorf("scan comment: %w", err)
		}
		comments = append(comments, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate comments: %w", err)
	}

	return comments, nil
}

// Create inserts a comment authored by userID on the given entity and
// returns it with the DB-generated fields (id, timestamps) and the author's
// display name resolved through the same users join List uses.
func (r *CommentsRepository) Create(ctx context.Context, userID, entityID string, entityType domain.CommentEntityType, body string) (domain.Comment, error) {
	// A comment belongs to exactly one entity; the FK column is part of the
	// static SQL, one query per entity type.
	var column string
	switch entityType {
	case domain.CommentEntityDot:
		column = "dot_id"
	case domain.CommentEntityObject:
		column = "object_id"
	case domain.CommentEntityPath:
		column = "path_id"
	default:
		return domain.Comment{}, fmt.Errorf("unknown comment entity type %q", entityType)
	}

	var c domain.Comment
	err := r.pool.QueryRow(ctx,
		`WITH ins AS (
			INSERT INTO comments (user_id, `+column+`, body) VALUES ($1, $2, $3)
			RETURNING id, user_id, created_at, modified_at
		)
		SELECT ins.id::text, u.first_name || ' ' || u.last_name, ins.created_at, ins.modified_at
		FROM ins JOIN users u ON u.id = ins.user_id`,
		userID, entityID, body).
		Scan(&c.ID, &c.Name, &c.CreatedAt, &c.ModifiedAt)
	if err != nil {
		return domain.Comment{}, fmt.Errorf("insert comment for entity %s: %w", entityID, err)
	}

	c.Body = body
	return c, nil
}
