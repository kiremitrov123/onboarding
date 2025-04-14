package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/kiremitrov123/onboarding/src/ogpreview/model"
	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	client *redis.Client
}

func NewClient(ctx context.Context, addr string) (*RedisClient, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
	})
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis at %s: %w", addr, err)
	}
	return &RedisClient{client: rdb}, nil
}

// GetTags gets the cached OG tags for a given URL.
// It returns an error if the value is not found or if unmarshalling fails.
func (rc *RedisClient) GetTags(url string) (*model.OGTags, error) {
	val, err := rc.client.Get(context.Background(), url).Result()
	if err != nil {
		return nil, err
	}
	var tags model.OGTags
	if err := json.Unmarshal([]byte(val), &tags); err != nil {
		return nil, err
	}
	return &tags, nil
}

// SetTags stores the given OG tags in Redis for the specified URL,
// with an expiry time of 10 minutes.
func (rc *RedisClient) SetTags(url string, tags model.OGTags) error {
	bytes, err := json.Marshal(tags)
	if err != nil {
		return err
	}
	return rc.client.Set(context.Background(), url, bytes, 10*time.Minute).Err()
}
