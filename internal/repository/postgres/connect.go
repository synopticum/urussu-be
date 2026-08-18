package postgres

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Connect creates a connection pool and waits for PostgreSQL to accept
// connections, giving a containerized database time to initialize.
func Connect(ctx context.Context, url string, log *slog.Logger) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	for attempt := 1; attempt <= 10; attempt++ {
		if err = pool.Ping(ctx); err == nil {
			return pool, nil
		}
		log.Info("database not ready, retrying", slog.Int("attempt", attempt), slog.Any("error", err))

		select {
		case <-ctx.Done():
			pool.Close()
			return nil, ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}

	pool.Close()
	return nil, fmt.Errorf("ping database: %w", err)
}

// RedactURL masks the password in a connection URL so it can be logged.
func RedactURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return "<invalid url>"
	}
	if _, hasPassword := u.User.Password(); hasPassword {
		u.User = url.UserPassword(u.User.Username(), "xxxxx")
	}
	return u.String()
}
