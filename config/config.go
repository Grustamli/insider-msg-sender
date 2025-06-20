// Package config loads and provides application configuration from environment variables
package config

import (
	"context"
	"strings"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/sethvargo/go-envconfig"
)

// Environment represents the running environment of the application (development or production).
type Environment string

const (
	// Development environment identifier
	Development Environment = "DEV"
	// Production environment identifier
	Production Environment = "PROD"
)

// AppConfig holds all application configuration settings sourced from environment variables.
// Fields include runtime environment, logging level, send intervals, and nested service configs.
type AppConfig struct {
	Environment             Environment    `env:"ENVIRONMENT, default=DEV"`              // run mode: DEV or PROD
	LogLevel                string         `env:"LOG_LEVEL, default=DEBUG"`              // verbosity level for logging
	SendIntervalSeconds     int            `env:"SEND_INTERVAL_SECONDS, default=120"`    // interval between send daemon runs
	MessageCountPerInterval int            `env:"MESSAGE_COUNT_PER_INTERVAL, default=2"` // messages to send per interval
	Postgres                PostgresConfig `env:", prefix=POSTGRES_"`                    // Postgres connection settings
	Webhook                 WebhookConfig  `env:", prefix=WEBHOOK_"`                     // Webhook sender settings
	Redis                   RedisConfig    `env:", prefix=REDIS_"`                       // Redis cache settings
}

// WebhookConfig holds HTTP webhook sender configuration options.
type WebhookConfig struct {
	URL            string `env:"URL"`                          // target webhook URL
	AuthHeader     string `env:"AUTH_HEADER"`                  // HTTP header name for auth key
	AuthKey        string `env:"AUTH_KEY"`                     // authentication key for webhook
	CharacterLimit int    `env:"CHARACTER_LIMIT, default=160"` // max message chars before truncation
	TimeoutSeconds int    `env:"TIMEOUT_SECONDS, default=20"`  // HTTP client timeout in seconds
}

// PostgresConfig holds the Postgres database connection URL.
type PostgresConfig struct {
	DBURL string `env:"DB_URL, required"` // Postgres DSN
}

// RedisConfig holds Redis client settings and cache key for message storage.
type RedisConfig struct {
	Address  string `env:"ADDRESS, default=localhost:6379"` // Redis server address
	DB       int    `env:"DB, default=0"`                   // Redis database number
	CacheKey string `env:"CACHE_KEY, default=messages"`     // key under which messages are cached
}

// IsProduction returns true if the configured environment is Production.
func (c *AppConfig) IsProduction() bool {
	return c.Environment == Production
}

// Load reads environment variables into an AppConfig instance,
// applying whitespace trimming to all values.
func Load(ctx context.Context) (*AppConfig, error) {
	ret := AppConfig{}
	// Process environment variables into the struct
	if err := envconfig.ProcessWith(ctx, &envconfig.Config{
		Target:   &ret,
		Mutators: []envconfig.Mutator{envconfig.MutatorFunc(trimConfigValue)},
	}); err != nil {
		return nil, errors.Wrap(err, "load config")
	}
	return &ret, nil
}

// trimConfigValue is an envconfig mutator that trims whitespace from values.
func trimConfigValue(_ context.Context, _, _, _, resolvedValue string) (string, bool, error) {
	return strings.TrimSpace(resolvedValue), false, nil
}

// Log outputs the loaded configuration at info level using the provided zerolog.Logger.
func (c *AppConfig) Log(l zerolog.Logger) {
	l.Info().Interface("config", c).Msg("Config")
}
