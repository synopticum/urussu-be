// Package config loads service configuration from a TOML file.
package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/BurntSushi/toml"
)

// Config is the root configuration structure, mapped from config.toml.
type Config struct {
	HTTP     HTTP     `toml:"http"`
	Database Database `toml:"database"`
	Log      Log      `toml:"log"`
}

type HTTP struct {
	Port int `toml:"port"`
}

type Database struct {
	URL string `toml:"url"`
	// SwaggerPath is the path to the generated OpenAPI schema,
	// served over HTTP at /swagger.json.
	SwaggerPath string `toml:"swagger_path"`
}

type Log struct {
	Level string `toml:"level"`
}

// Default returns the configuration used when no config file is present.
func Default() Config {
	return Config{
		HTTP: HTTP{Port: 8080},
		Database: Database{
			URL:         "postgres://urussu:urussu@localhost:5432/urussu_v2?sslmode=disable",
			SwaggerPath: "gen/openapiv2/urussu/v1/dots.swagger.json",
		},
		Log: Log{Level: "info"},
	}
}

// Load reads configuration from the TOML file at path, falling back to
// defaults when mustExist is false and the file is missing. A small set of
// environment variables (DATABASE_URL, HTTP_PORT) overrides the
// file values, which keeps container deployments simple.
func Load(path string, mustExist bool) (Config, error) {
	cfg := Default()

	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		if mustExist || !os.IsNotExist(err) {
			return Config{}, fmt.Errorf("load config %q: %w", path, err)
		}
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

	return cfg, nil
}
