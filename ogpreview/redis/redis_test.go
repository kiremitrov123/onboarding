package redis

import (
	"context"
	"testing"

	"github.com/kiremitrov123/onboarding/src/ogpreview/model"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func setupTestRedis(t *testing.T) *RedisClient {
	t.Helper()

	ctx := context.Background()
	rdb, err := NewClient(ctx, "localhost:6379")
	assert.NoError(t, err)

	err = rdb.client.FlushDB(ctx).Err()
	assert.NoError(t, err)

	return rdb
}

func TestSetAndGetTags(t *testing.T) {
	rdb := setupTestRedis(t)
	url := "http://test.com"

	tags := model.OGTags{
		Title:       "Test Title",
		Description: "Test Description",
		Image:       "http://test.com/image.jpg",
	}

	err := rdb.SetTags(url, tags)
	assert.NoError(t, err)

	cached, err := rdb.GetTags(url)
	assert.NoError(t, err)
	assert.Equal(t, tags.Title, cached.Title)
	assert.Equal(t, tags.Description, cached.Description)
	assert.Equal(t, tags.Image, cached.Image)
}

func TestGetTags_KeyNotSet(t *testing.T) {
	rdb := setupTestRedis(t)
	_, err := rdb.GetTags("http://not-set.com")
	assert.Error(t, err)
	assert.Equal(t, redis.Nil, err)
}
