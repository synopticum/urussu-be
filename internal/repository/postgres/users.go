package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"urussu-be/internal/domain"
)

// uniqueViolation is the SQLSTATE for a UNIQUE constraint breach
// (users.email on duplicate registration).
const uniqueViolation = "23505"

// UsersRepository reads and writes users in PostgreSQL.
type UsersRepository struct {
	pool *pgxpool.Pool
}

func NewUsersRepository(pool *pgxpool.Pool) *UsersRepository {
	return &UsersRepository{pool: pool}
}

func (r *UsersRepository) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	var u domain.User

	err := r.pool.QueryRow(ctx,
		`SELECT id::text, email, password, role, first_name, last_name
		 FROM users
		 WHERE email = $1`, email).
		Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.FirstName, &u.LastName)

	if errors.Is(err, pgx.ErrNoRows) {
		return domain.User{}, domain.ErrNotFound
	}

	if err != nil {
		return domain.User{}, fmt.Errorf("query user %s: %w", email, err)
	}

	return u, nil
}

func (r *UsersRepository) Create(ctx context.Context, user domain.User) (domain.User, error) {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO users (email, password, role, first_name, last_name)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id::text`,
		user.Email, user.PasswordHash, user.Role, user.FirstName, user.LastName).
		Scan(&user.ID)

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == uniqueViolation {
		return domain.User{}, domain.ErrAlreadyExists
	}

	if err != nil {
		return domain.User{}, fmt.Errorf("insert user %s: %w", user.Email, err)
	}

	return user, nil
}
