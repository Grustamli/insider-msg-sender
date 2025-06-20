// Package main initializes and runs the Insider Message Sender service,
// wiring together configuration, logging, repositories, sender, daemon, and API server.
package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"

	"github.com/grustamli/insider-msg-sender/api"
	"github.com/grustamli/insider-msg-sender/application"
	"github.com/grustamli/insider-msg-sender/config"
	"github.com/grustamli/insider-msg-sender/daemon"
	"github.com/grustamli/insider-msg-sender/logging"
	"github.com/grustamli/insider-msg-sender/message"
	"github.com/grustamli/insider-msg-sender/postgres"
	"github.com/grustamli/insider-msg-sender/postgres/gen"
	redisint "github.com/grustamli/insider-msg-sender/redis"
	"github.com/grustamli/insider-msg-sender/webhook"
)

// main is the entry point: it runs application startup and exits on error.
func main() {
	if err := run(); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

// run orchestrates loading config, initializing components, starting background tasks,
// and launching the HTTP API server.
func run() error {
	ctx := context.Background()

	// load application configuration
	cfg, err := config.Load(ctx)
	if err != nil {
		return err
	}

	// initialize structured logger
	log := initLogger(cfg)
	cfg.Log(log)

	// set up message repository (DB + Redis cache)
	messages, err := initMessageRepository(cfg)
	if err != nil {
		return err
	}

	// set up HTTP-based webhook sender
	sender, err := initMessageSender(cfg)
	if err != nil {
		return err
	}

	// wrap application with logging middleware
	app := logging.LogApplicationAccess(application.NewApplication(messages, sender), log)

	// send any unsent messages immediately
	go sendAllUnsentMessages(ctx, app, log)

	// start periodic daemon to send messages
	msgSenderDaemon := initMessageSenderDaemon(cfg, app, log)
	if err := msgSenderDaemon.Start(ctx); err != nil {
		return err
	}

	// initialize and run HTTP API server
	srv := initAPIServer(app, msgSenderDaemon, log)
	return srv.Run()
}

// sendAllUnsentMessages invokes SendAllUnsent and logs any error.
func sendAllUnsentMessages(ctx context.Context, app *logging.Application, log zerolog.Logger) {
	if err := app.SendAllUnsent(ctx); err != nil {
		log.Error().Err(err).Msg("Failed to send all unsent messages")
	}
}

// initLogger configures zerolog.Logger based on application settings.
func initLogger(cfg *config.AppConfig) zerolog.Logger {
	return logging.New(logging.LogConfig{
		IsProduction: cfg.IsProduction(),
		Level:        logging.Level(cfg.LogLevel),
	})
}

// initMessageRepository combines PostgreSQL storage and Redis caching for messages.
func initMessageRepository(cfg *config.AppConfig) (message.Repository, error) {
	// open Postgres connection
	db, err := initDB(cfg)
	if err != nil {
		return nil, err
	}

	// create Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.Redis.Address,
		DB:   cfg.Redis.DB,
	})

	// wrap the Postgres repo with Redis cache
	return redisint.NewCacheRepository(rdb, cfg.Redis.CacheKey,
		postgres.NewMessageRepository(gen.New(db)),
	), nil
}

// initDB opens a database/sql.DB connection to Postgres.
func initDB(cfg *config.AppConfig) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.Postgres.DBURL)
	if err != nil {
		return nil, errors.Wrap(err, "connecting to postgres db")
	}
	return db, nil
}

// initMessageSender constructs a webhook.MessageSender with timeouts and headers.
func initMessageSender(cfg *config.AppConfig) (*webhook.MessageSender, error) {
	client := &http.Client{Timeout: time.Duration(cfg.Webhook.TimeoutSeconds) * time.Second}
	sender, err := webhook.NewWebhookSender(client, cfg.Webhook.URL, buildWebhookOpts(&cfg.Webhook)...)
	if err != nil {
		return nil, errors.Wrap(err, "creating webhook sender")
	}
	return sender, nil
}

// buildWebhookOpts assembles functional options for the webhook sender.
func buildWebhookOpts(cfg *config.WebhookConfig) []webhook.OptFunc {
	var opts []webhook.OptFunc
	if cfg.CharacterLimit > 0 {
		opts = append(opts, webhook.WithCharacterLimit(cfg.CharacterLimit))
	}
	if cfg.AuthKey != "" {
		opts = append(opts, webhook.WithHeader(cfg.AuthHeader, cfg.AuthKey))
	}
	return opts
}

// initMessageSenderDaemon creates a TimerDaemon that sends a configured number
// of messages at regular intervals.
func initMessageSenderDaemon(cfg *config.AppConfig, app application.App, log zerolog.Logger) *daemon.TimerDaemon {
	return daemon.NewTimerDaemon("MessageSender", func(ctx context.Context) error {
		for i := 0; i < cfg.MessageCountPerInterval; i++ {
			if err := app.SendNext(ctx); err != nil {
				return err
			}
		}
		return nil
	}, time.Duration(cfg.SendIntervalSeconds)*time.Second, &log)
}

// initAPIServer constructs and returns the HTTP API server instance.
func initAPIServer(app application.App, msgSenderDaemon daemon.Daemon, log zerolog.Logger) *api.Server {
	return api.NewServer(gin.Default(), ":8000", app, msgSenderDaemon, log)
}
