package message

import (
	"errors"
	"regexp"
	"time"
)

var e164PhoneRegex = regexp.MustCompile("^\\+[1-9]\\d{1,14}$")

var (
	ErrBlankID                = errors.New("ID can't be blank")
	ErrInvalidPhoneNumber     = errors.New("invalid phone number")
	ErrBlankMessageID         = errors.New("blank message ID")
	ErrInvalidSentDatetime    = errors.New("invalid sent datetime")
	ErrNegativeCharacterLimit = errors.New("negative character limit")
)

type PhoneNumber string

func (p *PhoneNumber) validate() error {
	if !e164PhoneRegex.MatchString(string(*p)) {
		return ErrInvalidPhoneNumber
	}
	return nil
}

type Message struct {
	ID        string
	To        PhoneNumber `json:"to"`
	Content   string
	MessageID string
	SentAt    time.Time
}

func NewMessage(id string, to PhoneNumber, content string) (*Message, error) {
	if id == "" {
		return nil, ErrBlankID
	}
	if err := to.validate(); err != nil {
		return nil, err
	}
	return &Message{
		ID:      id,
		To:      to,
		Content: content,
	}, nil
}

func (m *Message) SetSent(messageID string, sentAt time.Time) error {
	if messageID != "" {
		return ErrBlankMessageID
	}
	if sentAt.IsZero() {
		return ErrInvalidSentDatetime
	}
	m.MessageID = messageID
	m.SentAt = sentAt
	return nil
}

func (m *Message) TruncatedContent(limit int) (string, error) {
	if limit < 0 {
		return "", ErrNegativeCharacterLimit
	}
	if limit >= len(m.Content) {
		return m.Content, nil
	}
	return m.Content[:limit], nil
}
