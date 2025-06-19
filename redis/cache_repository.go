package redis

import (
	"context"
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"insider-message-sender/message"
)

type CacheRepository struct {
	message.Repository
	rdb *redis.Client
	key string
}

var _ message.Repository = (*CacheRepository)(nil)

func NewCacheRepository(rdb *redis.Client, key string, repo message.Repository) *CacheRepository {
	return &CacheRepository{
		rdb:        rdb,
		key:        key,
		Repository: repo,
	}
}

func (c *CacheRepository) Save(ctx context.Context, msg *message.Message) error {
	if err := c.Repository.Save(ctx, msg); err != nil {
		return err
	}
	return c.saveMessageToCache(ctx, msg)
}

func (c *CacheRepository) GetAllSent(ctx context.Context) ([]*message.SentMessage, error) {
	// Query cache for the sent messages
	msgs, err := c.getMessagesFromCache(ctx)
	if err != nil {
		return nil, err
	}

	if len(msgs) > 0 {
		return msgs, nil
	}

	// Query repository when cache is empty
	msgs, err = c.Repository.GetAllSent(ctx)
	if err != nil {
		return nil, err
	}

	// cache messages retrieved from the db
	if err := c.saveAllToCache(ctx, msgs); err != nil {
		return nil, err
	}

	return msgs, nil
}

func (c *CacheRepository) saveMessageToCache(ctx context.Context, msg *message.Message) error {
	msgJson, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	if err := c.rdb.LPush(ctx, c.key, msgJson).Err(); err != nil {
		return errors.Wrap(err, "adding message to cache")
	}
	return nil
}

func (c *CacheRepository) saveAllToCache(ctx context.Context, msgs []*message.SentMessage) error {
	msgsJson, err := marshalMessages(msgs)
	if err != nil {
		return err
	}
	if err := c.rdb.LPush(ctx, c.key, msgsJson...).Err(); err != nil {
		return errors.Wrap(err, "adding messages to cache")
	}
	return nil
}

func (c *CacheRepository) getMessagesFromCache(ctx context.Context) ([]*message.SentMessage, error) {
	msgStrs, err := c.rdb.LRange(ctx, c.key, 0, -1).Result()
	if err != nil {
		return nil, errors.Wrap(err, "getting sent messages from cache")
	}
	return unmarshalMessageStrings(msgStrs)
}

func marshalMessages(msgs []*message.SentMessage) ([]any, error) {
	ret := make([]any, len(msgs))
	for i, m := range msgs {
		msgJson, err := json.Marshal(m)
		if err != nil {
			return nil, err
		}
		ret[i] = msgJson
	}
	return ret, nil
}

func unmarshalMessageStrings(strs []string) ([]*message.SentMessage, error) {
	ret := make([]*message.SentMessage, len(strs))
	for i, m := range strs {
		msg := &message.SentMessage{}
		if err := json.Unmarshal([]byte(m), msg); err != nil {
			return nil, err
		}
		ret[i] = msg
	}
	return ret, nil
}
