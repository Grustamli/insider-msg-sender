// Package main implements the CLI tool for seeding the Insider Message Sender database.
package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/alecthomas/kong"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/grustamli/insider-msg-sender/daemon"
	"github.com/grustamli/insider-msg-sender/logging"
	"github.com/grustamli/insider-msg-sender/message"
	"github.com/grustamli/insider-msg-sender/postgres"
	"github.com/grustamli/insider-msg-sender/postgres/gen"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

// MessageRepository defines an interface for inserting Message entities.
type MessageRepository interface {
	// Insert adds a new message to the repository.
	Insert(ctx context.Context, message *message.Message) error
}

// cli holds the top-level command definitions parsed by Kong.
var cli struct {
	Seed struct {
		DBURL    string `help:"Postgres Database URL (or set $DATABASE_URL)" env:"DATABASE_URL" name:"db-url"`
		Interval int    `short:"i" help:"Interval in seconds between seed runs. 0 = run once." default:"0"`
		Count    int    `short:"c" help:"Number of messages to insert each run. Default is 1" default:"1"`
	} `cmd help:"Seed the database with initial data."`
}

// main parses CLI arguments and dispatches to the appropriate command handler.
func main() {
	if err := run(); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

// run handles the chosen CLI command and returns an error if execution fails.
func run() error {
	// Parse command-line arguments into cli struct
	ctx := kong.Parse(&cli,
		kong.Name("cli"),
		kong.Description("Insider Message Sender CLI tool"),
		kong.UsageOnError(),
	)

	switch ctx.Command() {
	case "seed":
		// Execute the seed command
		if err := runSeed(); err != nil {
			return err
		}
	default:
		// Print usage for unknown commands
		return ctx.PrintUsage(false)
	}
	return nil
}

// runSeed initializes the repository and either seeds once or at intervals.
func runSeed() error {
	// Ensure a database URL is provided
	if cli.Seed.DBURL == "" {
		return errors.New("no database URL provided: set --db-url or $DATABASE_URL")
	}
	// Initialize the message repository
	messages, err := initMessageRepository(cli.Seed.DBURL)
	// Set up a console logger for the seeder
	logger := logging.New(logging.LogConfig{Level: logging.DEBUG})

	if err != nil {
		return err
	}
	ctx := context.Background()

	// Decide between single-run or periodic seeding
	if cli.Seed.Interval > 0 {
		return seedInIntervals(ctx, messages, cli.Seed.Interval, cli.Seed.Count, logger)
	}
	return seedMessages(ctx, messages, cli.Seed.Count)
}

// seedInIntervals starts a TimerDaemon that seeds messages at regular intervals.
// It blocks until the context is canceled, then stops the daemon gracefully.
func seedInIntervals(ctx context.Context, messages *postgres.MessageRepository, interval, count int, logger zerolog.Logger) error {
	// Create a new TimerDaemon for seeding
	d := daemon.NewTimerDaemon("MessageSeeder", func(ctx context.Context) error {
		return seedMessages(ctx, messages, count)
	}, time.Duration(interval)*time.Second, &logger)

	// Start the daemon
	if err := d.Start(ctx); err != nil {
		return err
	}

	// Wait until context cancellation (e.g., CTRL+C)
	<-ctx.Done()

	// Stop the daemon and clean up
	return d.Stop(context.Background())
}

// seedMessages inserts the specified number of fake messages into the repository.
func seedMessages(ctx context.Context, repo MessageRepository, count int) error {
	// Generate and insert messages
	if err := insertMessages(ctx, repo, createSeedMessages(count)); err != nil {
		return err
	}
	fmt.Println("Finished seeding messages")
	return nil
}

// initMessageRepository opens a Postgres connection and returns a MessageRepository.
func initMessageRepository(dsn string) (*postgres.MessageRepository, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	// Create a new Postgres-backed repository
	return postgres.NewMessageRepository(gen.New(db)), nil
}

// createSeedMessages generates a slice of fake Message objects for seeding.
// Each message has a randomized phone number and sentence content.
func createSeedMessages(count int) []*message.Message {
	ret := make([]*message.Message, count)
	for i := 0; i < count; i++ {
		ret[i] = &message.Message{
			To:      gofakeit.Numerify("+994#########"),
			Content: gofakeit.Sentence(6),
		}
	}
	return ret
}

// insertMessages writes each Message to the repository, returning on the first error.
func insertMessages(ctx context.Context, repo MessageRepository, messages []*message.Message) error {
	for _, msg := range messages {
		if err := insertMessage(ctx, repo, msg); err != nil {
			return errors.Wrap(err, "inserting message")
		}
	}
	return nil
}

// insertMessage writes individual message to the repository, returning error if failed
func insertMessage(ctx context.Context, repo MessageRepository, msg *message.Message) error {
	return repo.Insert(ctx, msg)
}
