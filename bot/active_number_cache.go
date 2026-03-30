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

func (c *ActiveNumberCache) activeKey(normalizedNumber string) string {
	return fmt.Sprintf("%s:active:number:%s", c.keyPrefix, normalizedNumber)
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
	activeKey := c.activeKey(normalized)

	pipe := c.client.Pipeline()
	pipe.Set(ctx, activeKey, payload, c.ttl)
	pipe.SAdd(ctx, userSetKey, normalized)
	pipe.Expire(ctx, userSetKey, c.ttl)
	_, err = pipe.Exec(ctx)
	return err
}

func (c *ActiveNumberCache) GetByNumber(ctx context.Context, number string) (*activenumber.ActiveNumber, error) {
	normalized := c.NormalizeNumber(number)
	if normalized == "" {
		return nil, nil
	}

	data, err := c.client.Get(ctx, c.activeKey(normalized)).Bytes()
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

func (c *ActiveNumberCache) DeleteByNumber(ctx context.Context, number string) error {
	normalized := c.NormalizeNumber(number)
	if normalized == "" {
		return nil
	}

	an, err := c.GetByNumber(ctx, number)
	if err != nil {
		return err
	}

	activeKey := c.activeKey(normalized)
	if an != nil {
		pipe := c.client.Pipeline()
		pipe.Del(ctx, activeKey)
		pipe.SRem(ctx, c.userSetKey(an.UserID), normalized)
		_, err = pipe.Exec(ctx)
		return err
	}

	return c.client.Del(ctx, activeKey).Err()
}

func (c *ActiveNumberCache) DeleteByUser(ctx context.Context, userID string) error {
	userKey := c.userSetKey(userID)
	numbers, err := c.client.SMembers(ctx, userKey).Result()
	if err != nil && err != redis.Nil {
		return err
	}

	pipe := c.client.Pipeline()
	for _, n := range numbers {
		pipe.Del(ctx, c.activeKey(n))
	}
	pipe.Del(ctx, userKey)
	_, err = pipe.Exec(ctx)
	return err
}
