package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://urussu:urussu@localhost:5432/urussu?sslmode=disable"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	conn, err := connectWithRetry(ctx, dsn)
	if err != nil {
		log.Fatalf("could not connect to database: %v", err)
	}
	defer conn.Close(ctx)

	fmt.Println("Connected to PostgreSQL.")

	// Sample read query #1: row counts per table.
	tables := []string{"collection", "comments", "dots", "objects", "paths", "users"}
	for _, table := range tables {
		var count int
		err := conn.QueryRow(ctx, fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&count)
		if err != nil {
			log.Fatalf("count query failed for %s: %v", table, err)
		}
		fmt.Printf("%-12s %d rows\n", table, count)
	}

	// Sample read query #2: a few dot titles.
	fmt.Println("\nSample dots:")
	rows, err := conn.Query(ctx, `SELECT id, title FROM dots LIMIT 5`)
	if err != nil {
		log.Fatalf("sample query failed: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id, title string
		if err := rows.Scan(&id, &title); err != nil {
			log.Fatalf("scan failed: %v", err)
		}
		fmt.Printf("  %s  %s\n", id, title)
	}
	if err := rows.Err(); err != nil {
		log.Fatalf("rows error: %v", err)
	}
}

// connectWithRetry gives the database container time to finish initializing.
func connectWithRetry(ctx context.Context, dsn string) (*pgx.Conn, error) {
	var err error
	for attempt := 1; attempt <= 10; attempt++ {
		var conn *pgx.Conn
		conn, err = pgx.Connect(ctx, dsn)
		if err == nil {
			if err = conn.Ping(ctx); err == nil {
				return conn, nil
			}
			conn.Close(ctx)
		}
		fmt.Printf("database not ready (attempt %d/10), retrying...\n", attempt)
		time.Sleep(2 * time.Second)
	}
	return nil, err
}
