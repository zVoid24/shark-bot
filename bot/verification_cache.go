package bot

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type VerificationCache struct {
	client    *redis.Client
	keyPrefix string
	ttl       time.Duration
}

func NewVerificationCache(client *redis.Client, keyPrefix string, ttl time.Duration) *VerificationCache {
	if ttl <= 0 {
		ttl = 2 * time.Hour
	}
	return &VerificationCache{
		client:    client,
		keyPrefix: keyPrefix,
		ttl:       ttl,
	}
}

func (c *VerificationCache) key(userID int64) string {
	return fmt.Sprintf("%s:verify:%d", c.keyPrefix, userID)
}

func (c *VerificationCache) IsVerified(ctx context.Context, userID int64) bool {
	if c.client == nil {
		return false
	}
	val, err := c.client.Get(ctx, c.key(userID)).Result()
	if err != nil {
		return false
	}
	return val == "1"
}

func (c *VerificationCache) SetVerified(ctx context.Context, userID int64) error {
	if c.client == nil {
		return nil
	}
	return c.client.Set(ctx, c.key(userID), "1", c.ttl).Err()
}

func (c *VerificationCache) Clear(ctx context.Context, userID int64) error {
	if c.client == nil {
		return nil
	}
	return c.client.Del(ctx, c.key(userID)).Err()
}
