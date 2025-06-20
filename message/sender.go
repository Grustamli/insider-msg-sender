package message

import (
	"context"
	"time"
)

// SendResult holds metadata about a successfully sent message.
// MessageID is the external provider's identifier for the message,
// SentAt is the timestamp when the message was sent.
type SendResult struct {
	MessageID string    // external provider message identifier
	SentAt    time.Time // timestamp when the message was sent
}

// Sender represents a service capable of sending Message entities.
// Implementations should handle delivery via an external provider and
// return a SendResult containing the provider-assigned ID and send time.
type Sender interface {
	// Send attempts to deliver the provided Message.
	// On success, it returns a SendResult and a nil error.
	// On failure, it returns a non-nil error.
	Send(ctx context.Context, msg *Message) (*SendResult, error)
}
