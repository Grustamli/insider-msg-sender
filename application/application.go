package application

import (
	"context"
	"github.com/pkg/errors"
	"insider-message-sender/message"
)

type App interface {
	SendNext(ctx context.Context) error
	SendAllUnsent(ctx context.Context) error
	ListSentMessages(ctx context.Context) ([]*message.SentMessage, error)
}

type Application struct {
	messages message.Repository
	sender   message.MessageSender
}

var _ App = (*Application)(nil)

func NewApplication(messages message.Repository, sender message.MessageSender) *Application {
	return &Application{
		messages: messages,
		sender:   sender,
	}
}

func (a *Application) SendNext(ctx context.Context) error {
	msg, err := a.messages.GetNextUnsent(ctx)
	if err != nil {
		return errors.Wrap(err, "getting next unsent message")
	}
	return a.sendMessage(ctx, msg)
}

func (a *Application) SendAllUnsent(ctx context.Context) error {
	msgs, err := a.messages.GetAllUnsent(ctx)
	if err != nil {
		return errors.Wrap(err, "getting all unsent messages")
	}
	for _, msg := range msgs {
		if err := a.sendMessage(ctx, msg); err != nil {
			return err
		}
	}
	return nil
}

func (a *Application) sendMessage(ctx context.Context, msg *message.Message) error {
	res, err := a.sender.Send(ctx, msg)
	if err != nil {
		return errors.Wrap(err, "sending message")
	}
	if err := msg.SetSent(res.MessageID, res.SentAt); err != nil {
		return errors.Wrap(err, "setting message sent status")
	}
	return a.messages.Save(ctx, msg)
}

func (a *Application) ListSentMessages(ctx context.Context) ([]*message.SentMessage, error) {
	ret, err := a.messages.GetAllSent(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "listing sent messages")
	}
	return ret, nil
}
