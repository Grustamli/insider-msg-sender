package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/grustamli/insider-msg-sender/message"
	"github.com/grustamli/insider-msg-sender/postgres/gen"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	"strconv"
)

type MessageRepository struct {
	queries *gen.Queries
}

var _ message.Repository = (*MessageRepository)(nil)

func NewMessageRepository(queries *gen.Queries) *MessageRepository {
	return &MessageRepository{
		queries: queries,
	}
}

func (m *MessageRepository) GetNextUnsent(ctx context.Context) (*message.Message, error) {
	res, err := m.queries.GetNextUnsent(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, errors.Wrap(err, "getting next unsent message")
	}
	return messageFromRow(res)
}

func messageFromRow(res gen.GetNextUnsentRow) (*message.Message, error) {
	return message.NewMessage(strID(res.ID), res.Recipient, res.Content)
}

func strID(id int32) string {
	return fmt.Sprintf("%d", id)
}

func (m *MessageRepository) Save(ctx context.Context, msg *message.Message) error {
	// if message is not set sent don't do any action
	if msg.SentAt.IsZero() {
		return nil
	}
	// otherwise assert that message id is not blank
	if msg.MessageID == "" {
		return errors.New("message ID is empty")
	}
	id, err := strconv.Atoi(msg.ID)
	if err != nil {
		return errors.Wrap(err, "converting message ID to int")
	}
	err = m.queries.SetMessageSent(ctx, gen.SetMessageSentParams{
		ID:        int32(id),
		SentAt:    sql.NullTime{Time: msg.SentAt, Valid: true},
		MessageID: sql.NullString{String: msg.MessageID, Valid: true},
	})
	if err != nil {
		return errors.Wrap(err, "setting message sent")
	}
	return nil
}

func (m *MessageRepository) GetAllSent(ctx context.Context) ([]*message.SentMessage, error) {
	res, err := m.queries.GetAllSent(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, errors.Wrap(err, "getting all sent message")
	}
	return sentMessagesFromRows(res)
}

func (m *MessageRepository) Insert(ctx context.Context, msg *message.Message) error {
	if err := m.queries.InsertMessage(ctx, gen.InsertMessageParams{
		Recipient: msg.To,
		Content:   msg.Content,
	}); err != nil {
		return errors.Wrap(err, "inserting message")
	}
	return nil
}

func sentMessagesFromRows(res []gen.GetAllSentRow) ([]*message.SentMessage, error) {
	ret := make([]*message.SentMessage, len(res))
	for i, r := range res {
		msg, err := sentMessageFromRow(r)
		if err != nil {
			return nil, err
		}
		ret[i] = msg
	}
	return ret, nil
}

func sentMessageFromRow(r gen.GetAllSentRow) (*message.SentMessage, error) {
	if !r.SentAt.Valid {
		return nil, fmt.Errorf("invalid sent timestamp, %v", r.SentAt.Time)
	}
	if !r.MessageID.Valid {
		return nil, fmt.Errorf("invalid message ID, %s", r.MessageID.String)
	}
	return &message.SentMessage{
		MessageID: r.MessageID.String,
		SentAt:    r.SentAt.Time,
	}, nil
}

func (m *MessageRepository) GetAllUnsent(ctx context.Context) ([]*message.Message, error) {
	res, err := m.queries.GetAllUnsent(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, errors.Wrap(err, "getting all unsent message")
	}
	return unsentMessagesFromRows(res)
}

func unsentMessagesFromRows(res []gen.GetAllUnsentRow) ([]*message.Message, error) {
	ret := make([]*message.Message, len(res))
	for i, r := range res {
		msg, err := message.NewMessage(strID(r.ID), r.Recipient, r.Content)
		if err != nil {
			return nil, errors.Wrap(err, "creating message from row")
		}
		ret[i] = msg
	}
	return ret, nil
}
