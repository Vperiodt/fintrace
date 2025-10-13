package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config aggregates application configuration values.
type Config struct {
	HTTP    HTTPConfig
	Graph   GraphConfig
	Logging LoggingConfig
}

// HTTPConfig governs HTTP server behaviour.
type HTTPConfig struct {
	Host              string
	Port              int
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	ShutdownTimeout   time.Duration
	MetricsEnabled    bool
	AllowedOriginsCSV string
}

// GraphConfig describes connectivity to the graph database (Neptune/Neo4j).
type GraphConfig struct {
	URI            string
	Database       string
	Username       string
	Password       string
	MaxConnections int
}

// LoggingConfig controls structured logging settings.
type LoggingConfig struct {
	Level         string
	Format        string // text|json
	Colored       bool
	IncludeCaller bool
}

const (
	defaultHost             = "0.0.0.0"
	defaultPort             = 8080
	defaultReadTimeout      = 10 * time.Second
	defaultWriteTimeout     = 15 * time.Second
	defaultIdleTimeout      = 60 * time.Second
	defaultShutdownTimeout  = 10 * time.Second
	defaultLoggingLevel     = "info"
	defaultLoggingFormat    = "text"
	defaultGraphMaxSessions = 10
)

// Load reads configuration from environment variables, applying defaults.
func Load() (Config, error) {
	cfg := Config{
		HTTP: HTTPConfig{
			Host:            valueOrDefault("SERVER_HOST", defaultHost),
			ReadTimeout:     defaultReadTimeout,
			WriteTimeout:    defaultWriteTimeout,
			IdleTimeout:     defaultIdleTimeout,
			ShutdownTimeout: defaultShutdownTimeout,
		},
		Logging: LoggingConfig{
			Level:         valueOrDefault("LOG_LEVEL", defaultLoggingLevel),
			Format:        valueOrDefault("LOG_FORMAT", defaultLoggingFormat),
			Colored:       parseBoolWithDefault("LOG_COLOR", false),
			IncludeCaller: parseBoolWithDefault("LOG_INCLUDE_CALLER", false),
		},
		Graph: GraphConfig{
			URI:            os.Getenv("GRAPH_URI"),
			Database:       valueOrDefault("GRAPH_DATABASE", ""),
			Username:       os.Getenv("GRAPH_USERNAME"),
			Password:       os.Getenv("GRAPH_PASSWORD"),
			MaxConnections: parseIntWithDefault("GRAPH_MAX_CONNECTIONS", defaultGraphMaxSessions),
		},
	}

	port, err := parsePort("SERVER_PORT", defaultPort)
	if err != nil {
		return Config{}, err
	}
	cfg.HTTP.Port = port

	if v := os.Getenv("SERVER_READ_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.HTTP.ReadTimeout = d
		} else {
			return Config{}, fmt.Errorf("invalid SERVER_READ_TIMEOUT: %w", err)
		}
	}

	if v := os.Getenv("SERVER_WRITE_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.HTTP.WriteTimeout = d
		} else {
			return Config{}, fmt.Errorf("invalid SERVER_WRITE_TIMEOUT: %w", err)
		}
	}

	if v := os.Getenv("SERVER_IDLE_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.HTTP.IdleTimeout = d
		} else {
			return Config{}, fmt.Errorf("invalid SERVER_IDLE_TIMEOUT: %w", err)
		}
	}

	if v := os.Getenv("SERVER_SHUTDOWN_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.HTTP.ShutdownTimeout = d
		} else {
			return Config{}, fmt.Errorf("invalid SERVER_SHUTDOWN_TIMEOUT: %w", err)
		}
	}

	cfg.HTTP.MetricsEnabled = parseBoolWithDefault("SERVER_METRICS_ENABLED", false)
	cfg.HTTP.AllowedOriginsCSV = os.Getenv("SERVER_ALLOWED_ORIGINS")

	return cfg, nil
}

func valueOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func parseBoolWithDefault(key string, fallback bool) bool {
	if v := os.Getenv(key); v != "" {
		val, err := strconv.ParseBool(v)
		if err != nil {
			return fallback
		}
		return val
	}
	return fallback
}

func parseIntWithDefault(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if val, err := strconv.Atoi(v); err == nil {
			return val
		}
	}
	return fallback
}

func parsePort(key string, fallback int) (int, error) {
	if v := os.Getenv(key); v != "" {
		port, err := strconv.Atoi(v)
		if err != nil {
			return 0, fmt.Errorf("invalid %s value %q: %w", key, v, err)
		}
		if port <= 0 || port > 65535 {
			return 0, fmt.Errorf("port %d is out of range", port)
		}
		return port, nil
	}
	return fallback, nil
}
