package cache

import (
	"context"
	"testing"
	"time"

	"github.com/kiremitrov123/onboarding/stockprice/model"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestRedisCache_SetGet(t *testing.T) {
	ctx := context.Background()
	rc, err := NewRedisCache(ctx, "localhost:6379")
	if err != nil {
		t.Fatalf("failed to connect to redis: %v", err)
	}

	symbol := "AAPL"
	price := model.Price{
		Symbol:    symbol,
		Price:     999.99,
		Timestamp: time.Now(),
	}

	err = rc.Set(ctx, symbol, price)
	assert.NoError(t, err)

	got, err := rc.Get(ctx, symbol)
	assert.NoError(t, err)
	assert.Equal(t, price.Symbol, got.Symbol)
	assert.Equal(t, price.Price, got.Price)
}

func TestRedisCache_SubscribeInvalidation(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rc, err := NewRedisCache(ctx, "localhost:6379")
	if err != nil {
		t.Fatalf("failed to connect to redis: %v", err)
	}

	key := "AAPL"
	invalidated := make(chan string, 1)

	err = rc.SubscribeInvalidation(ctx, func(k string) {
		invalidated <- k
	})
	assert.NoError(t, err)

	// Give the subscriber goroutine a moment to subscribe
	time.Sleep(200 * time.Millisecond)

	// Publish an invalidation message manually
	err = redis.NewClient(&redis.Options{Addr: "localhost:6379"}).
		Publish(ctx, "__redis__:invalidate", key).Err()
	assert.NoError(t, err)

	select {
	case msg := <-invalidated:
		assert.Equal(t, key, msg)
	case <-time.After(1 * time.Second):
		t.Error("expected invalidation callback to be triggered")
	}
}
