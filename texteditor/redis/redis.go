package redis

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kiremitrov123/onboarding/texteditor/model"
	"github.com/redis/go-redis/v9"
)

type PubSub interface {
	PublishEdit(ctx context.Context, docID string, edit model.Edit) error
	SubscribeEdits(ctx context.Context, docID string, handler func(model.Edit)) error
}

type RedisClient struct {
	client *redis.Client
}

func NewRedisClient(ctx context.Context, addr string) (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}
	return &RedisClient{client: client}, nil
}

func (rc *RedisClient) PublishEdit(ctx context.Context, docID string, edit model.Edit) error {
	data, err := json.Marshal(edit)
	if err != nil {
		return fmt.Errorf("failed to marshal edit: %w", err)
	}
	return rc.client.Publish(ctx, docID, data).Err()
}

func (rc *RedisClient) SubscribeEdits(ctx context.Context, docID string, handler func(model.Edit)) error {
	sub := rc.client.Subscribe(ctx, docID)
	ch := sub.Channel()

	go func() {
		for msg := range ch {
			var edit model.Edit
			if err := json.Unmarshal([]byte(msg.Payload), &edit); err == nil {
				handler(edit)
			}
		}
	}()

	return nil
}
