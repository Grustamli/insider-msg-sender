package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/alecthomas/kong"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"insider-message-sender/daemon"
	"insider-message-sender/logging"
	"insider-message-sender/message"
	"insider-message-sender/postgres"
	"insider-message-sender/postgres/gen"
	"os"
	"time"
)

type MessageRepository interface {
	Insert(ctx context.Context, message *message.Message) error
}

var cli struct {
	Seed struct {
		DBURL    string `help:"Postgres DSN (or set $DATABASE_URL)" env:"DATABASE_URL" name:"db-url"`
		Interval int    `short:"i" help:"Interval in seconds between seed runs. 0 = run once." default:"0"`
		Count    int    `short:"c" help:"Number of messages to insert each run. Default is 1" default:"1"`
	} `cmd help:"Seed the database with initial data."`
}

func main() {
	if err := run(); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func run() error {
	ctx := kong.Parse(&cli,
		kong.Name("cli"),
		kong.Description("Insider Message Sender CLI tool"),
		kong.UsageOnError(),
	)

	switch ctx.Command() {
	case "seed":
		if err := runSeed(); err != nil {
			return err
		}
	default:
		return ctx.PrintUsage(false)
	}
	return nil
}

func runSeed() error {
	dsn := cli.Seed.DBURL
	if dsn == "" {
		return errors.New("no database URL provided: set --db-url or $DATABASE_URL")
	}
	messages, err := initMessageRepository(dsn)
	logger := logging.New(logging.LogConfig{Level: logging.DEBUG})

	if err != nil {
		return err
	}
	ctx := context.Background()

	if cli.Seed.Interval > 0 {
		return seedInIntervals(ctx, messages, cli.Seed.Interval, cli.Seed.Count, logger)
	}
	return seedMessages(ctx, messages, cli.Seed.Count)
}

func seedInIntervals(ctx context.Context, messages *postgres.MessageRepository, interval, count int, logger zerolog.Logger) error {
	d := daemon.NewTimerDaemon("MessageSeeder", func(ctx context.Context) error {
		return seedMessages(ctx, messages, count)
	}, time.Duration(interval)*time.Second, &logger)

	if err := d.Start(ctx); err != nil {
		return err
	}

	// Wait for context cancellation
	<-ctx.Done()

	// Stop the daemon gracefully
	return d.Stop(context.Background())
}

func seedMessages(ctx context.Context, repo MessageRepository, count int) error {
	if err := insertMessages(ctx, repo, createSeedMessages(count)); err != nil {
		return err
	}
	fmt.Println("Finished seeding messages")
	return nil
}

func initMessageRepository(dsn string) (*postgres.MessageRepository, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	return postgres.NewMessageRepository(gen.New(db)), nil
}

func createSeedMessages(count int) []*message.Message {
	ret := make([]*message.Message, count)
	for i := 0; i < count; i++ {
		ret[i] = &message.Message{
			To:      message.PhoneNumber(gofakeit.Numerify("+994#########")),
			Content: gofakeit.Sentence(6),
		}
	}
	return ret
}

func insertMessages(ctx context.Context, repo MessageRepository, messages []*message.Message) error {
	for _, msg := range messages {
		if err := repo.Insert(ctx, msg); err != nil {
			return errors.Wrap(err, "inserting message")
		}
	}
	return nil
}
