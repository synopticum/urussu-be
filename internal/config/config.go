// Package config provides service configuration: built-in defaults embedded
// from config.env (the single source shared with the Makefile and
// docker-compose), with environment variable overrides on top.
package config

import (
	_ "embed"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
)

//go:embed config.env
var defaultEnv []byte

// Config is the root configuration structure.
type Config struct {
	HTTP     HTTP
	Database Database
	Log      Log
}

type HTTP struct {
	Port int
}

type Database struct {
	URL string
}

type Log struct {
	Level string
}

// Load returns the default configuration with environment variable
// overrides applied: HTTP_PORT, DATABASE_URL, LOG_LEVEL.
func Load() (Config, error) {
	cfg, err := defaults()
	if err != nil {
		return Config{}, err
	}

	if v, ok := os.LookupEnv("DATABASE_URL"); ok {
		cfg.Database.URL = v
	}
	if v, ok := os.LookupEnv("HTTP_PORT"); ok {
		port, err := strconv.Atoi(v)
		if err != nil {
			return Config{}, fmt.Errorf("invalid HTTP_PORT %q: %w", v, err)
		}
		cfg.HTTP.Port = port
	}
	if v, ok := os.LookupEnv("LOG_LEVEL"); ok {
		cfg.Log.Level = v
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

// defaults parses the embedded config.env. The database URL is assembled
// from the POSTGRES_* parts with a localhost host (local development);
// docker-compose assembles its own URL with the db host from the same parts.
func defaults() (Config, error) {
	vars := parseDotenv(string(defaultEnv))

	for _, key := range []string{"HTTP_PORT", "LOG_LEVEL", "POSTGRES_USER", "POSTGRES_PASSWORD", "POSTGRES_DB", "POSTGRES_PORT"} {
		if vars[key] == "" {
			return Config{}, fmt.Errorf("config.env: %s is required", key)
		}
	}

	port, err := strconv.Atoi(vars["HTTP_PORT"])
	if err != nil {
		return Config{}, fmt.Errorf("config.env: invalid HTTP_PORT %q: %w", vars["HTTP_PORT"], err)
	}
	if _, err := strconv.Atoi(vars["POSTGRES_PORT"]); err != nil {
		return Config{}, fmt.Errorf("config.env: invalid POSTGRES_PORT %q: %w", vars["POSTGRES_PORT"], err)
	}

	return Config{
		HTTP: HTTP{Port: port},
		Database: Database{
			URL: fmt.Sprintf("postgres://%s:%s@localhost:%s/%s?sslmode=disable",
				vars["POSTGRES_USER"], vars["POSTGRES_PASSWORD"], vars["POSTGRES_PORT"], vars["POSTGRES_DB"]),
		},
		Log: Log{Level: vars["LOG_LEVEL"]},
	}, nil
}

// parseDotenv parses KEY=VALUE lines, ignoring blank lines and # comments.
func parseDotenv(data string) map[string]string {
	vars := make(map[string]string)
	for line := range strings.Lines(data) {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if k, v, ok := strings.Cut(line, "="); ok {
			vars[strings.TrimSpace(k)] = strings.TrimSpace(v)
		}
	}
	return vars
}

// Validate checks that the configuration is usable, so the service fails
// fast on startup instead of misbehaving later.
func (c Config) Validate() error {
	if c.HTTP.Port < 1 || c.HTTP.Port > 65535 {
		return fmt.Errorf("http port %d out of range 1-65535", c.HTTP.Port)
	}
	if c.Database.URL == "" {
		return errors.New("database url must not be empty")
	}
	var lvl slog.Level
	if err := lvl.UnmarshalText([]byte(c.Log.Level)); err != nil {
		return fmt.Errorf("invalid log level %q: %w", c.Log.Level, err)
	}
	return nil
}
