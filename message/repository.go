package message

import (
	"context"
	"time"
)

// SentMessage represents a record of a successfully sent message.
// It includes the external provider's message ID and the timestamp when it was sent.
type SentMessage struct {
	MessageID string    `json:"message_id"` // external provider message identifier
	SentAt    time.Time `json:"sent_at"`    // timestamp when the message was sent
}

// Repository provides methods to store and retrieve messages from a data store.
// It supports fetching unsent and sent messages, as well as updating send status.
type Repository interface {
	// GetNextUnsent returns the next Message that has not yet been sent.
	// If there are no unsent messages, it returns (nil, nil).
	GetNextUnsent(ctx context.Context) (*Message, error)

	// GetAllUnsent returns all Messages that are not yet sent.
	// Returns an empty slice or nil if no unsent messages exist.
	GetAllUnsent(ctx context.Context) ([]*Message, error)

	// GetAllSent returns all SentMessage records for messages that have been sent.
	// Returns an empty slice or nil if no sent messages exist.
	GetAllSent(ctx context.Context) ([]*SentMessage, error)

	// Save updates the repository with the provided Message's sent state.
	// It should persist the MessageID and SentAt timestamp.
	// Returns an error if the update fails.
	Save(ctx context.Context, msg *Message) error
}

// RepositoryMiddleware defines a decorator that wraps a Repository with additional behavior.
type RepositoryMiddleware func(Repository) Repository

// RepositoryWithMiddleware applies one or more RepositoryMiddleware decorators to a base Repository.
// Middleware is applied in reverse order, so the first argument wraps the second, and so on.
func RepositoryWithMiddleware(repo Repository, mws ...RepositoryMiddleware) Repository {
	r := repo
	// apply middleware in reverse to ensure correct ordering
	for i := len(mws) - 1; i >= 0; i-- {
		r = mws[i](r)
	}
	return r
}
