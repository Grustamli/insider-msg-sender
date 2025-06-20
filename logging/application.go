package logging

import (
	"context"
	"github.com/grustamli/insider-msg-sender/application"
	"github.com/grustamli/insider-msg-sender/message"
	"github.com/rs/zerolog"
)

type Application struct {
	application.App
	logger zerolog.Logger
}

func LogApplicationAccess(application application.App, logger zerolog.Logger) *Application {
	return &Application{
		App:    application,
		logger: logger,
	}
}

func (a *Application) SendNext(ctx context.Context) (err error) {
	a.logger.Info().Msg("--> Application.SendNext")
	defer func() { a.logger.Info().Err(err).Msg("<-- Application.SendNext") }()
	return a.App.SendNext(ctx)
}

func (a *Application) SendAllUnsent(ctx context.Context) (err error) {
	a.logger.Info().Msg("--> Application.SendAllUnsent")
	defer func() { a.logger.Info().Err(err).Msg("<-- Application.SendAllUnsent") }()
	return a.App.SendAllUnsent(ctx)
}

func (a *Application) ListSentMessages(ctx context.Context) (msgs []*message.SentMessage, err error) {
	a.logger.Info().Msg("--> Application.ListSentMessages")
	defer func() { a.logger.Info().Err(err).Msg("<-- Application.ListSentMessages") }()
	return a.App.ListSentMessages(ctx)
}
