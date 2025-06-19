package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"insider-message-sender/api"
	"insider-message-sender/application"
	"insider-message-sender/config"
	"insider-message-sender/daemon"
	"insider-message-sender/logging"
	"insider-message-sender/message"
	"insider-message-sender/postgres"
	"insider-message-sender/postgres/gen"
	redisint "insider-message-sender/redis"
	"insider-message-sender/webhook"
	"net/http"
	"os"
	"time"
)

func main() {
	if err := run(); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func run() error {
	ctx := context.Background()

	cfg, err := config.Load(ctx)
	if err != nil {
		return err
	}
	log := initLogger(cfg)
	cfg.Log(log)

	messages, err := initMessageRepository(cfg)
	if err != nil {
		return err
	}

	sender, err := initMessageSender(cfg)
	if err != nil {
		return err
	}

	app := logging.LogApplicationAccess(application.NewApplication(messages, sender), log)

	if err := app.SendAllUnsent(ctx); err != nil {
		return err
	}

	msgSenderDaemon := initMessageSenderDaemon(cfg, app, log)
	if err := msgSenderDaemon.Start(ctx); err != nil {
		return err
	}
	srv := initAPIServer(app, msgSenderDaemon)
	return srv.Run()
}

func initLogger(cfg *config.AppConfig) zerolog.Logger {
	return logging.New(logging.LogConfig{
		IsProduction: cfg.IsProduction(),
		Level:        logging.Level(cfg.LogLevel),
	})
}
func initMessageRepository(cfg *config.AppConfig) (message.Repository, error) {
	db, err := initDB(cfg)

	if err != nil {
		return nil, err
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.Redis.Address,
		DB:   cfg.Redis.DB,
	})

	return redisint.NewCacheRepository(rdb, cfg.Redis.CacheKey,
		postgres.NewMessageRepository(gen.New(db)),
	), nil
}

func initDB(cfg *config.AppConfig) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.Postgres.DBURL)
	if err != nil {
		return nil, errors.Wrap(err, "connecting to postgres db")
	}
	return db, nil
}

func initMessageSender(cfg *config.AppConfig) (*webhook.MessageSender, error) {
	client := initHTTPClient(&cfg.Webhook)
	ret, err := webhook.NewWebhookSender(client, cfg.Webhook.URL, buildWebhookOpts(&cfg.Webhook)...)
	if err != nil {
		return nil, errors.Wrap(err, "creating webhook sender")
	}
	return ret, nil
}

func buildWebhookOpts(cfg *config.WebhookConfig) []webhook.OptFunc {
	var ret []webhook.OptFunc

	if cfg.CharacterLimit > 0 {
		ret = append(ret, webhook.WithCharacterLimit(cfg.CharacterLimit))
	}
	if cfg.AuthKey != "" {
		ret = append(ret, webhook.WithHeader(cfg.AuthHeader, cfg.AuthKey))
	}
	return ret
}

func initHTTPClient(cfg *config.WebhookConfig) *http.Client {
	client := &http.Client{Timeout: time.Second * time.Duration(cfg.TimeoutSeconds)}
	return client
}

func initMessageSenderDaemon(cfg *config.AppConfig, app application.App, logger zerolog.Logger) *daemon.TimerDaemon {
	return daemon.NewTimerDaemon("MessageSender", func(ctx context.Context) error {
		for i := 0; i < cfg.MessageCountPerInterval; i++ {
			if err := app.SendNext(ctx); err != nil {
				return err
			}
		}
		return nil
	}, time.Duration(cfg.SendIntervalSeconds)*time.Second, &logger)
}

func initAPIServer(app application.App, msgSenderDaemon daemon.Daemon) *api.Server {
	return api.NewServer(gin.Default(), ":8000", app, msgSenderDaemon)
}
