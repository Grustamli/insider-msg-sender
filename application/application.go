// Package application defines the core business logic for sending messages.
// It composes a message.Repository for persistence and a message.Sender for delivery.
package application

import (
	"context"
	"time"

	"github.com/grustamli/insider-msg-sender/message"
	"github.com/pkg/errors"
)

// App defines the operations available for sending messages.
// - SendNext sends the next unsent message, if one exists.
// - SendAllUnsent sends all pending unsent messages.
// - ListSentMessages returns all messages that have already been sent.
type App interface {
	// SendNext retrieves and sends a single unsent message.
	// Returns nil if there are no unsent messages.
	SendNext(ctx context.Context) error

	// SendAllUnsent retrieves and sends all unsent messages.
	// It pauses for one second between each send to avoid burst traffic.
	SendAllUnsent(ctx context.Context) error

	// ListSentMessages returns all sent messages recorded in the system.
	ListSentMessages(ctx context.Context) ([]*message.SentMessage, error)
}

// Application is the default implementation of the App interface.
// It uses a message.Repository to manage message state and a message.Sender to deliver messages.
type Application struct {
	messages message.Repository // repository for message persistence
	sender   message.Sender     // sender for delivering messages
}

var _ App = (*Application)(nil) // assert Application implements App

// NewApplication constructs a new Application with the provided repository and sender.
func NewApplication(messages message.Repository, sender message.Sender) *Application {
	return &Application{
		messages: messages,
		sender:   sender,
	}
}

// SendNext retrieves the next unsent message from the repository and sends it.
// If no unsent message is found, it returns without error.
// Any errors fetching or sending are wrapped and returned.
func (a *Application) SendNext(ctx context.Context) error {
	msg, err := a.messages.GetNextUnsent(ctx)
	if err != nil {
		return errors.Wrap(err, "getting next unsent message")
	}
	if msg == nil {
		// nothing to send
		return nil
	}
	return a.sendMessage(ctx, msg)
}

// SendAllUnsent retrieves all unsent messages and sends them one by one.
// It sleeps for one second between sends to throttle the rate.
// Errors during retrieval or send abort the process immediately.
func (a *Application) SendAllUnsent(ctx context.Context) error {
	msgs, err := a.messages.GetAllUnsent(ctx)
	if err != nil {
		return errors.Wrap(err, "getting all unsent messages")
	}
	for _, msg := range msgs {
		if err := a.sendMessage(ctx, msg); err != nil {
			return err
		}
		// brief pause to avoid overwhelming sender
		time.Sleep(time.Second)
	}
	return nil
}

// sendMessage executes the delivery of a single message, marks it as sent, and persists the update.
// Returns any errors encountered during send or save operations.
func (a *Application) sendMessage(ctx context.Context, msg *message.Message) error {
	res, err := a.sender.Send(ctx, msg)
	if err != nil {
		return errors.Wrap(err, "sending message")
	}
	// update message state with external ID and timestamp
	if err := msg.SetSent(res.MessageID, res.SentAt); err != nil {
		return errors.Wrap(err, "setting message sent status")
	}
	return a.messages.Save(ctx, msg)
}

// ListSentMessages retrieves all messages marked as sent from the repository.
// Errors during retrieval are wrapped and returned.
func (a *Application) ListSentMessages(ctx context.Context) ([]*message.SentMessage, error) {
	ret, err := a.messages.GetAllSent(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "listing sent messages")
	}
	return ret, nil
}
