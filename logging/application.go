package logging

import (
	"context"

	"github.com/grustamli/insider-msg-sender/application"
	"github.com/grustamli/insider-msg-sender/message"
	"github.com/rs/zerolog"
)

// Application wraps an application.App instance with logging middleware.
// It logs calls to the SendNext, SendAllUnsent, and ListSentMessages methods.
type Application struct {
	application.App                // embedded application interface
	logger          zerolog.Logger // logger to record method invocations
}

// LogApplicationAccess returns a new logging.Application that wraps the given App
// and emits log entries using the provided zerolog.Logger.
func LogApplicationAccess(app application.App, logger zerolog.Logger) *Application {
	return &Application{
		App:    app,
		logger: logger,
	}
}

// SendNext logs entry and exit for the SendNext method and delegates to the underlying App.
// It logs an info message before and after the call, including any error.
func (a *Application) SendNext(ctx context.Context) (err error) {
	a.logger.Info().Msg("--> Application.SendNext")
	defer func() { a.logger.Info().Err(err).Msg("<-- Application.SendNext") }()
	return a.App.SendNext(ctx)
}

// SendAllUnsent logs entry and exit for the SendAllUnsent method and delegates to the underlying App.
// It logs an info message before and after the call, including any error.
func (a *Application) SendAllUnsent(ctx context.Context) (err error) {
	a.logger.Info().Msg("--> Application.SendAllUnsent")
	defer func() { a.logger.Info().Err(err).Msg("<-- Application.SendAllUnsent") }()
	return a.App.SendAllUnsent(ctx)
}

// ListSentMessages logs entry and exit for the ListSentMessages method and delegates to the underlying App.
// It logs an info message before and after the call, including returned messages and any error.
func (a *Application) ListSentMessages(ctx context.Context) (msgs []*message.SentMessage, err error) {
	a.logger.Info().Msg("--> Application.ListSentMessages")
	defer func() { a.logger.Info().Err(err).Msg("<-- Application.ListSentMessages") }()
	return a.App.ListSentMessages(ctx)
}
