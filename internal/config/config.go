// Package config provides service configuration: built-in defaults with
// environment variable overrides.
package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
)

// Config is the root configuration structure.
type Config struct {
	HTTP     HTTP
	Database Database
	Swagger  Swagger
	Log      Log
}

type HTTP struct {
	Port int
}

type Database struct {
	URL string
}

type Swagger struct {
	// Path is the path to the generated OpenAPI schema,
	// served over HTTP at /swagger.json.
	Path string
}

type Log struct {
	Level string
}

// Default returns the built-in configuration used for local development.
func Default() Config {
	return Config{
		HTTP: HTTP{Port: 8080},
		Database: Database{
			URL: "postgres://urussu:urussu@localhost:5432/urussu_v2?sslmode=disable",
		},
		Swagger: Swagger{
			Path: "gen/openapiv2/urussu/v1/dots.swagger.json",
		},
		Log: Log{Level: "info"},
	}
}

// Load returns the default configuration with environment variable
// overrides applied: HTTP_PORT, DATABASE_URL, SWAGGER_PATH, LOG_LEVEL.
func Load() (Config, error) {
	cfg := Default()

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
	if v, ok := os.LookupEnv("SWAGGER_PATH"); ok {
		cfg.Swagger.Path = v
	}
	if v, ok := os.LookupEnv("LOG_LEVEL"); ok {
		cfg.Log.Level = v
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

// Validate checks that the configuration is usable, so the service fails
// fast on startup instead of misbehaving later.
func (c Config) Validate() error {
	if c.HTTP.Port < 1 || c.HTTP.Port > 65535 {
		return fmt.Errorf("http port %d out of range 1-65535", c.HTTP.Port)
	}
	if c.Database.URL == "" {
		return fmt.Errorf("database url must not be empty")
	}
	var lvl slog.Level
	if err := lvl.UnmarshalText([]byte(c.Log.Level)); err != nil {
		return fmt.Errorf("invalid log level %q: %w", c.Log.Level, err)
	}
	return nil
}
