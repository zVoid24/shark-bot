package bot

//testing
import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"shark_bot/internal/activenumber"

	"github.com/redis/go-redis/v9"
)

var cacheDigitOnly = regexp.MustCompile(`\D`)

type ActiveNumberCache struct {
	client    *redis.Client
	keyPrefix string
	ttl       time.Duration
}

func NewActiveNumberCache(client *redis.Client, keyPrefix string, ttl time.Duration) *ActiveNumberCache {
	pfx := strings.TrimSpace(keyPrefix)
	if pfx == "" {
		pfx = "sharkbot"
	}
	if ttl <= 0 {
		ttl = 2 * time.Hour
	}
	return &ActiveNumberCache{
		client:    client,
		keyPrefix: pfx,
		ttl:       ttl,
	}
}

func (c *ActiveNumberCache) NormalizeNumber(number string) string {
	return cacheDigitOnly.ReplaceAllString(number, "")
}

func (c *ActiveNumberCache) activeKey(normalizedNumber, platform string) string {
	return fmt.Sprintf("%s:active:number:%s:plat:%s", c.keyPrefix, normalizedNumber, platform)
}

func (c *ActiveNumberCache) userSetKey(userID string) string {
	return fmt.Sprintf("%s:active:user:%s", c.keyPrefix, userID)
}

func (c *ActiveNumberCache) Set(ctx context.Context, an activenumber.ActiveNumber) error {
	normalized := c.NormalizeNumber(an.Number)
	if normalized == "" {
		return fmt.Errorf("active cache set failed: empty normalized number")
	}
	payload, err := json.Marshal(an)
	if err != nil {
		return err
	}

	userSetKey := c.userSetKey(an.UserID)
	activeKey := c.activeKey(normalized, an.Platform)

	pipe := c.client.Pipeline()
	pipe.Set(ctx, activeKey, payload, c.ttl)
	// We store "normalized:platform" in the user's set
	pipe.SAdd(ctx, userSetKey, fmt.Sprintf("%s:%s", normalized, an.Platform))
	pipe.Expire(ctx, userSetKey, c.ttl)
	_, err = pipe.Exec(ctx)
	return err
}

func (c *ActiveNumberCache) GetByNumber(ctx context.Context, number, platform string) (*activenumber.ActiveNumber, error) {
	normalized := c.NormalizeNumber(number)
	if normalized == "" {
		return nil, nil
	}

	data, err := c.client.Get(ctx, c.activeKey(normalized, platform)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var an activenumber.ActiveNumber
	if err := json.Unmarshal(data, &an); err != nil {
		return nil, err
	}
	return &an, nil
}

func (c *ActiveNumberCache) DeleteByNumber(ctx context.Context, number, platform string) error {
	normalized := c.NormalizeNumber(number)
	if normalized == "" {
		return nil
	}

	activeKey := c.activeKey(normalized, platform)
	pipe := c.client.Pipeline()
	pipe.Del(ctx, activeKey)
	// We don't easily know the user here to remove from SRem without a lookup, 
	// but user set will eventually expire or be cleaned up in DeleteByUser
	_, err := pipe.Exec(ctx)
	return err
}

func (c *ActiveNumberCache) DeleteByUser(ctx context.Context, userID string) error {
	userKey := c.userSetKey(userID)
	entries, err := c.client.SMembers(ctx, userKey).Result()
	if err != nil && err != redis.Nil {
		return err
	}

	pipe := c.client.Pipeline()
	for _, entry := range entries {
		// entry is "normalized:platform"
		parts := strings.Split(entry, ":")
		if len(parts) == 2 {
			pipe.Del(ctx, c.activeKey(parts[0], parts[1]))
		} else {
			// fallback for old format keys
			pipe.Del(ctx, fmt.Sprintf("%s:active:number:%s", c.keyPrefix, entry))
		}
	}
	pipe.Del(ctx, userKey)
	_, err = pipe.Exec(ctx)
	return err
}
