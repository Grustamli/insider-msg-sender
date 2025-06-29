// Package redis provides a caching decorator for message.Repository implementations
// using Redis lists to store and retrieve sent message metadata.
package redis

import (
	"context"
	"encoding/json"

	"github.com/grustamli/insider-msg-sender/message"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
)

// CacheRepository wraps a message.Repository and adds Redis-based caching
// for sent messages under a specified key.
// It delegates unsent operations to the underlying repository.
type CacheRepository struct {
	message.Repository               // underlying repository for persistence
	rdb                *redis.Client // Redis client instance
	key                string        // Redis list key for caching sent messages
}

var _ message.Repository = (*CacheRepository)(nil) // ensure interface compliance

// NewCacheRepository constructs a CacheRepository that uses rdb and key for caching,
// delegating other operations to repo.
func NewCacheRepository(rdb *redis.Client, key string, repo message.Repository) *CacheRepository {
	return &CacheRepository{
		rdb:        rdb,
		key:        key,
		Repository: repo,
	}
}

// Save persists the message status via the underlying repository
// and then caches the sent message metadata in Redis.
func (c *CacheRepository) Save(ctx context.Context, msg *message.Message) error {
	if err := c.Repository.Save(ctx, msg); err != nil {
		return err
	}
	return c.saveMessageToCache(ctx, msg)
}

// GetAllSent returns all sent messages from cache if present;
// otherwise, it falls back to the underlying repository, caches the results, then returns them.
func (c *CacheRepository) GetAllSent(ctx context.Context) ([]*message.SentMessage, error) {
	// attempt to read from cache
	msgs, err := c.getMessagesFromCache(ctx)
	if err != nil {
		return nil, err
	}
	if len(msgs) > 0 {
		return msgs, nil
	}
	// cache miss: query underlying repository
	msgs, err = c.Repository.GetAllSent(ctx)
	if err != nil {
		return nil, err
	}
	// populate cache for future calls
	if err := c.saveAllToCache(ctx, msgs); err != nil {
		return nil, err
	}
	return msgs, nil
}

// saveMessageToCache serializes a single SentMessage and pushes it onto the Redis list.
func (c *CacheRepository) saveMessageToCache(ctx context.Context, msg *message.Message) error {
	data, err := json.Marshal(&message.SentMessage{MessageID: msg.MessageID, SentAt: msg.SentAt})
	if err != nil {
		return err
	}
	if err := c.rdb.LPush(ctx, c.key, data).Err(); err != nil {
		return errors.Wrap(err, "adding message to cache")
	}
	return nil
}

// saveAllToCache serializes multiple SentMessage entries and pushes them all onto the Redis list.
func (c *CacheRepository) saveAllToCache(ctx context.Context, msgs []*message.SentMessage) error {
	items, err := marshalMessages(msgs)
	if err != nil {
		return err
	}
	if err := c.rdb.LPush(ctx, c.key, items...).Err(); err != nil {
		return errors.Wrap(err, "adding messages to cache")
	}
	return nil
}

// getMessagesFromCache reads all entries from the Redis list and deserializes them into SentMessage objects.
func (c *CacheRepository) getMessagesFromCache(ctx context.Context) ([]*message.SentMessage, error) {
	entries, err := c.rdb.LRange(ctx, c.key, 0, -1).Result()
	if err != nil {
		return nil, errors.Wrap(err, "getting sent messages from cache")
	}
	return unmarshalMessageStrings(entries)
}

// marshalMessages serializes each SentMessage into JSON for Redis storage.
func marshalMessages(msgs []*message.SentMessage) ([]any, error) {
	ret := make([]any, len(msgs))
	for i, m := range msgs {
		data, err := json.Marshal(m)
		if err != nil {
			return nil, err
		}
		ret[i] = data
	}
	return ret, nil
}

// unmarshalMessageStrings converts JSON strings from Redis into SentMessage objects.
func unmarshalMessageStrings(strs []string) ([]*message.SentMessage, error) {
	ret := make([]*message.SentMessage, len(strs))
	for i, s := range strs {
		var msg message.SentMessage
		if err := json.Unmarshal([]byte(s), &msg); err != nil {
			return nil, err
		}
		ret[i] = &msg
	}
	return ret, nil
}
