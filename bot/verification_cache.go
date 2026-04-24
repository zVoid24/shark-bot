package bot

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type VerificationCache struct {
	client    *redis.Client
	keyPrefix string
	ttl       time.Duration
	recheck   time.Duration
	memCache  sync.Map // fallback if redis is nil
}

func NewVerificationCache(client *redis.Client, keyPrefix string, ttl time.Duration) *VerificationCache {
	if ttl <= 0 {
		ttl = 2 * time.Hour
	}
	recheck := 2 * time.Minute
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

type memEntry struct {
	verified  bool
	timestamp time.Time
}

func (c *VerificationCache) key(userID int64) string {
	return fmt.Sprintf("%s:verify:%d", c.keyPrefix, userID)
}

func (c *VerificationCache) IsVerified(ctx context.Context, userID int64) bool {
	// 1. Check Memory Cache (Always check first for speed)
	if val, ok := c.memCache.Load(userID); ok {
		entry := val.(memEntry)
		if entry.verified && time.Since(entry.timestamp) < c.recheck {
			return true
		}
	}

	// 2. Check Redis if available
	if c.client != nil {
		key := c.key(userID)
		val, err := c.client.Get(ctx, key).Result()
		if err == nil && val == "1" {
			// Double check TTL for recheck logic
			ttl, err := c.client.TTL(ctx, key).Result()
			if err == nil && ttl > 0 {
				age := c.ttl - ttl
				if age < c.recheck {
					// Backfill memCache and return
					c.memCache.Store(userID, memEntry{verified: true, timestamp: time.Now().Add(-age)})
					return true
				}
			}
		}
	}

	return false
}

// IsUnverifiedRecently returns true if the user was recently checked and found unverified.
func (c *VerificationCache) IsUnverifiedRecently(ctx context.Context, userID int64) bool {
	if val, ok := c.memCache.Load(userID); ok {
		entry := val.(memEntry)
		if !entry.verified && time.Since(entry.timestamp) < 1*time.Minute {
			return true
		}
	}
	return false
}

func (c *VerificationCache) SetVerified(ctx context.Context, userID int64) error {
	// Store in memory
	c.memCache.Store(userID, memEntry{verified: true, timestamp: time.Now()})

	// Store in Redis if available
	if c.client != nil {
		return c.client.Set(ctx, c.key(userID), "1", c.ttl).Err()
	}
	return nil
}

func (c *VerificationCache) SetUnverified(ctx context.Context, userID int64) {
	// Only store in memory for short duration (negative cache)
	c.memCache.Store(userID, memEntry{verified: false, timestamp: time.Now()})
}

func (c *VerificationCache) Clear(ctx context.Context, userID int64) error {
	c.memCache.Delete(userID)
	if c.client != nil {
		return c.client.Del(ctx, c.key(userID)).Err()
	}
	return nil
}
