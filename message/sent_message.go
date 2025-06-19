package message

import "time"

type SentMessage struct {
	MessageID string    `json:"message_id"`
	SentAt    time.Time `json:"sent_at"`
}
