// Package message defines the core Message entity and related validation logic.
package message

import (
	"errors"
	"regexp"
	"time"
)

var (
	// e164PhoneRegex matches valid E.164 phone number format (e.g., +1234567890).
	e164PhoneRegex = regexp.MustCompile("^\\+[1-9]\\d{1,14}$")
)

var (
	// ErrBlankID is returned when attempting to create a Message without an ID.
	ErrBlankID = errors.New("ID can't be blank")

	// ErrInvalidPhoneNumber is returned when the recipient phone number is not E.164-compliant.
	ErrInvalidPhoneNumber = errors.New("invalid phone number")

	// ErrBlankMessageID is returned when setting the sent state without a message ID.
	ErrBlankMessageID = errors.New("blank message ID")

	// ErrInvalidSentDatetime is returned when setting the sent state with a zero timestamp.
	ErrInvalidSentDatetime = errors.New("invalid sent datetime")

	// ErrNegativeCharacterLimit is returned when truncating content with a negative limit.
	ErrNegativeCharacterLimit = errors.New("negative character limit")
)

// validatePhone ensures the given number matches E.164 format.
func validatePhone(num string) error {
	if !e164PhoneRegex.MatchString(num) {
		return ErrInvalidPhoneNumber
	}
	return nil
}

// Message represents an outbound message with recipient information and send metadata.
// ID is the internal identifier, To is the E.164 phone number, Content is the message body.
type Message struct {
	ID        string    // internal message identifier
	To        string    // recipient phone number in E.164 format
	Content   string    // message payload
	MessageID string    // external message provider ID after sending
	SentAt    time.Time // timestamp when the message was sent
}

// NewMessage constructs a new Message with the given id, recipient, and content.
// Returns ErrBlankID if id is empty, or ErrInvalidPhoneNumber if To is invalid.
func NewMessage(id, to, content string) (*Message, error) {
	if id == "" {
		return nil, ErrBlankID
	}
	if err := validatePhone(to); err != nil {
		return nil, err
	}
	return &Message{
		ID:      id,
		To:      to,
		Content: content,
	}, nil
}

// SetSent marks the Message as sent by providing an external messageID and sentAt timestamp.
// Returns ErrBlankMessageID if messageID is empty, or ErrInvalidSentDatetime if sentAt is zero.
func (m *Message) SetSent(messageID string, sentAt time.Time) error {
	if messageID == "" {
		return ErrBlankMessageID
	}
	if sentAt.IsZero() {
		return ErrInvalidSentDatetime
	}
	m.MessageID = messageID
	m.SentAt = sentAt
	return nil
}

// TruncatedContent returns the Content truncated to at most limit characters.
// If limit is negative, returns ErrNegativeCharacterLimit.
// If limit >= len(Content), returns the full Content.
func (m *Message) TruncatedContent(limit int) (string, error) {
	if limit < 0 {
		return "", ErrNegativeCharacterLimit
	}
	if limit >= len(m.Content) {
		return m.Content, nil
	}
	return m.Content[:limit], nil
}
