package config

import (
	"context"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/sethvargo/go-envconfig"
	"strings"
)

type Environment string

const (
	Development Environment = "DEV"
	Production  Environment = "PROD"
)

type AppConfig struct {
	Environment         Environment    `env:"ENVIRONMENT,default=DEV"`
	LogLevel            string         `env:"LOG_LEVEL,default=DEBUG"`
	SendIntervalSeconds int            `env:"SEND_INTERVAL_SECONDS,default=120"`
	Postgres            PostgresConfig `env:", prefix=POSTGRES_"`
	Webhook             WebhookConfig  `env:", prefix=WEBHOOK_"`
	Redis               RedisConfig    `env:", prefix=REDIS_"`
}

type WebhookConfig struct {
	URL            string `env:"URL"`
	AuthHeader     string `env:"AUTH_HEADER"`
	AuthKey        string `env:"AUTH_KEY"`
	CharacterLimit int    `env:"CHARACTER_LIMIT, default=160"` // typical character limit for SMS as default
	TimeoutSeconds int    `env:"TIMEOUT_SECONDS, default=20"`
}

type PostgresConfig struct {
	DBURL string `env:"DB_URL, required"`
}

type RedisConfig struct {
	Address  string `env:"ADDRESS, default=localhost:6379"`
	DB       int    `env:"DB, default=0"`
	CacheKey string `env:"CACHE_KEY, default=messages"`
}

func (c *AppConfig) IsProduction() bool {
	return c.Environment == Production
}

func Load(ctx context.Context) (*AppConfig, error) {
	ret := AppConfig{}
	if err := envconfig.ProcessWith(ctx, &envconfig.Config{
		Target: &ret,
		Mutators: []envconfig.Mutator{
			envconfig.MutatorFunc(trimConfigValue),
		},
	}); err != nil {
		return nil, errors.Wrap(err, "load config")
	}
	return &ret, nil
}

func trimConfigValue(_ context.Context, _, _, _, resolvedValue string) (string, bool, error) {
	return strings.TrimSpace(resolvedValue), false, nil
}

func (c *AppConfig) Log(l zerolog.Logger) {
	l.Info().Interface("config", c).Msg("Config")
}
