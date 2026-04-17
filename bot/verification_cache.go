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
	recheck   time.Duration
}

func NewVerificationCache(client *redis.Client, keyPrefix string, ttl time.Duration) *VerificationCache {
	if ttl <= 0 {
		ttl = 2 * time.Hour
	}
	recheck := 10 * time.Minute
	if ttl < recheck {
		recheck = ttl
	}
	return &VerificationCache{
		client:    client,
		keyPrefix: keyPrefix,
		ttl:       ttl,
		recheck:   recheck,
	}
}

func (c *VerificationCache) key(userID int64) string {
	return fmt.Sprintf("%s:verify:%d", c.keyPrefix, userID)
}

func (c *VerificationCache) IsVerified(ctx context.Context, userID int64) bool {
	if c.client == nil {
		return false
	}
	key := c.key(userID)
	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		return false
	}

	if val != "1" {
		return false
	}

	// Defensive check: only trust cache entries that have a live TTL.
	// If TTL is missing (TTL=-1), treat it as stale and force Telegram re-check.
	ttl, err := c.client.TTL(ctx, key).Result()
	if err != nil || ttl <= 0 {
		return false
	}

	// Force periodic revalidation so leaving a group is picked up quickly.
	age := c.ttl - ttl
	if c.recheck > 0 && age >= c.recheck {
		return false
	}

	return true
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
