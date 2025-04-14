package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/kiremitrov123/onboarding/stockprice/model"
	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(ctx context.Context, addr string) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	// Enable client-side caching with BCAST on the main client connection
	if err := client.Do(ctx, "CLIENT", "TRACKING", "ON", "BCAST").Err(); err != nil {
		return nil, fmt.Errorf("failed to enable CLIENT TRACKING: %w", err)
	}

	// Verify connectivity
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &RedisCache{client: client}, nil
}

func (rc *RedisCache) Get(ctx context.Context, symbol string) (*model.Price, error) {
	val, err := rc.client.Get(ctx, symbol).Result()
	if err != nil {
		return nil, err
	}
	var price model.Price
	if err := json.Unmarshal([]byte(val), &price); err != nil {
		return nil, err
	}
	return &price, nil
}

func (rc *RedisCache) Set(ctx context.Context, symbol string, price model.Price) error {
	bytes, err := json.Marshal(price)
	if err != nil {
		return err
	}

	if err := rc.client.Set(ctx, symbol, bytes, 10*time.Minute).Err(); err != nil {
		return err
	}

	rc.client.Publish(ctx, "__redis__:invalidate", symbol)

	return nil
}

func (rc *RedisCache) SubscribeInvalidation(ctx context.Context, onInvalidate func(key string)) error {
	pubsub := rc.client.PSubscribe(ctx, "__redis__:invalidate")

	go func() {
		for msg := range pubsub.Channel() {
			onInvalidate(msg.Payload)
		}
	}()

	return nil
}
