package message

import (
	"context"
	"time"
)

type SendResult struct {
	MessageID string
	SentAt    time.Time
}

type MessageSender interface {
	Send(ctx context.Context, msg *Message) (*SendResult, error)
}
