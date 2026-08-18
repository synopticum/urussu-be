package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/jackc/pgx/v5/pgxpool"

	"urussu-be/gen/openapiv2"
	urussuv1 "urussu-be/gen/urussu/v1"
	"urussu-be/internal/config"
	"urussu-be/internal/handlers/grpc"
	"urussu-be/internal/repository/postgres"
	"urussu-be/internal/service"
)

func main() {
	if err := run(); err != nil {
		slog.Error("fatal", slog.Any("error", err))
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	log := newLogger(cfg.Log.Level)
	slog.SetDefault(log)

	log.Info("config loaded",
		slog.Int("http_port", cfg.HTTP.Port),
		slog.String("log_level", cfg.Log.Level),
		slog.String("database_url", redactURL(cfg.Database.URL)))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := connectDB(ctx, cfg.Database.URL, log)
	if err != nil {
		return err
	}
	defer pool.Close()
	log.Info("connected to PostgreSQL")

	// HTTP server: grpc-gateway handlers run in-process (no gRPC listener),
	// plus the generated OpenAPI schema.
	httpMux := http.NewServeMux()

	gwMux := runtime.NewServeMux()
	dotsRepo := postgres.NewDotsRepository(pool)
	dotsService := service.NewDotsService(dotsRepo)
	if err := urussuv1.RegisterDotsServiceHandlerServer(ctx, gwMux, grpc.NewDotsHandler(dotsService, log)); err != nil {
		return fmt.Errorf("register dots gateway: %w", err)
	}
	httpMux.Handle("/api/", gwMux)

	httpMux.Handle("GET /swagger.json", swaggerHandler())

	httpAddr := fmt.Sprintf(":%d", cfg.HTTP.Port)
	httpServer := &http.Server{
		Addr:              httpAddr,
		Handler:           httpMux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Info("HTTP server listening", slog.String("addr", httpAddr))
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("http server: %w", err)
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
	}

	log.Info("shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("http shutdown: %w", err)
	}

	return nil
}

func newLogger(level string) *slog.Logger {
	var lvl slog.Level
	// Cannot fail: the level is validated in config.Load.
	_ = lvl.UnmarshalText([]byte(level))
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: lvl}))
}

// redactURL masks the password in a connection URL so it can be logged.
func redactURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return "<invalid url>"
	}
	if _, hasPassword := u.User.Password(); hasPassword {
		u.User = url.UserPassword(u.User.Username(), "xxxxx")
	}
	return u.String()
}

// connectDB creates a connection pool and waits for PostgreSQL to accept
// connections, giving a containerized database time to initialize.
func connectDB(ctx context.Context, url string, log *slog.Logger) (*pgxpool.Pool, error) {
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

// swaggerHandler serves the OpenAPI schema embedded from the generated
// gen/openapiv2 output (see gen/openapiv2/embed.go).
func swaggerHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(openapiv2.SwaggerSchema)
	})
}
