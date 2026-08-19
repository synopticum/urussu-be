package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/cors"

	urussuv1 "urussu-be/gen/urussu/v1"
	"urussu-be/internal/auth"
	"urussu-be/internal/config"
	"urussu-be/internal/handlers/grpc"
	"urussu-be/internal/handlers/swagger"
	"urussu-be/internal/repository/postgres"
	"urussu-be/internal/service"
)

func main() {
	if err := run(); err != nil {
		// run() may have failed before the configured logger was installed
		// (e.g. config.Load), so log through an equivalent one to keep the
		// output format consistent.
		newLogger(slog.LevelInfo.String()).Error("fatal", slog.Any("error", err))
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
		slog.Any("cors_allowed_origins", cfg.HTTP.CORSAllowedOrigins),
		slog.String("database_url", postgres.RedactURL(cfg.Database.URL)))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := postgres.Connect(ctx, cfg.Database.URL, log)
	if err != nil {
		return err
	}
	defer pool.Close()
	log.Info("connected to PostgreSQL")

	// HTTP server: grpc-gateway handlers run in-process (no gRPC listener),
	// plus the generated OpenAPI schema.
	httpMux := http.NewServeMux()
	gwMux := runtime.NewServeMux()

	// Services
	dotsRepo := postgres.NewDotsRepository(pool)
	dotsService := service.NewDotsService(dotsRepo)
	if err := urussuv1.RegisterDotsServiceHandlerServer(ctx, gwMux, grpc.NewDotsHandler(dotsService, log)); err != nil {
		return fmt.Errorf("register dots gateway: %w", err)
	}

	// Auth
	usersRepo := postgres.NewUsersRepository(pool)
	authService := service.NewAuthService(usersRepo, cfg.JWT.Secret)
	if err := urussuv1.RegisterAuthServiceHandlerServer(ctx, gwMux, grpc.NewAuthHandler(authService, log)); err != nil {
		return fmt.Errorf("register auth gateway: %w", err)
	}
	// The auth endpoints must stay reachable without a token; everything
	// else under /api/ goes through JWT authentication. The secret itself
	// is never logged.
	log.Info("JWT authentication enabled", slog.String("public_prefix", "/api/v1/auth/"))
	httpMux.Handle("/api/", auth.Middleware(cfg.JWT.Secret, log, "/api/v1/auth/")(gwMux))

	// Swagger
	httpMux.Handle("GET /swagger.json", swagger.JsonHandler())
	httpMux.Handle("GET /swagger/", swagger.UIHandler())

	// CORS is needed only when the frontend is served from a different
	// origin; an empty origin list leaves the mux unwrapped (rs/cors would
	// otherwise default to allowing every origin).
	var handler http.Handler = httpMux
	if len(cfg.HTTP.CORSAllowedOrigins) > 0 {
		handler = cors.New(cors.Options{
			AllowedOrigins: cfg.HTTP.CORSAllowedOrigins,
			AllowedMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
			AllowedHeaders: []string{"Content-Type", "Authorization"},
			MaxAge:         600,
		}).Handler(httpMux)
	}

	httpAddr := fmt.Sprintf(":%d", cfg.HTTP.Port)
	httpServer := &http.Server{
		Addr:              httpAddr,
		Handler:           handler,
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
