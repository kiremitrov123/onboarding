package redis

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kiremitrov123/onboarding/commenting/model"
	"github.com/stretchr/testify/require"
)

func setupRedis(t *testing.T) *RedisCache {
	ctx := context.Background()
	cache, err := NewCache(ctx, "localhost:6379")
	require.NoError(t, err)

	// Clear test keys under "comments:*"
	keys, err := cache.client.Keys(ctx, "comments:*").Result()
	require.NoError(t, err)

	if len(keys) > 0 {
		_, err := cache.client.Del(ctx, keys...).Result()
		require.NoError(t, err)
	}

	return cache
}

func TestSetAndGetComment(t *testing.T) {
	ctx := context.Background()
	cache := setupRedis(t)

	c := model.Comment{
		ID:        uuid.New(),
		ThreadID:  uuid.New(),
		UserID:    "user123",
		Content:   "This is a comment",
		CreatedAt: time.Now(),
		Upvotes:   10,
	}

	err := cache.SetComment(ctx, &c)
	require.NoError(t, err)

	fetched, err := cache.GetCommentByID(ctx, c.ID)
	require.NoError(t, err)
	require.Equal(t, c.ID, fetched.ID)
	require.Equal(t, c.Content, fetched.Content)
	require.Equal(t, c.Upvotes, fetched.Upvotes)
}

func TestListComments_CacheHit(t *testing.T) {
	ctx := context.Background()
	cache := setupRedis(t)

	threadID := uuid.New()
	c := model.Comment{
		ID:         uuid.New(),
		ThreadID:   threadID,
		UserID:     "user123",
		Content:    "From Redis",
		CreatedAt:  time.Now(),
		ReplyCount: 5,
		Upvotes:    7,
	}

	// Store in Redis
	err := cache.SetComment(ctx, &c)
	require.NoError(t, err)

	comments, err := cache.ListComments(ctx, threadID, "upvotes", int64(c.Upvotes+1), 10, func(context.Context, uuid.UUID) ([]model.Comment, error) {
		t.Fatal("should not call fallback")
		return nil, nil
	})
	require.NoError(t, err)
	require.Len(t, comments, 1)
	require.Equal(t, c.ID, comments[0].ID)
}

func TestListComments_CacheMissFallback(t *testing.T) {
	ctx := context.Background()
	cache := setupRedis(t)

	threadID := uuid.New()
	c := model.Comment{
		ID:         uuid.New(),
		ThreadID:   threadID,
		UserID:     "user123",
		Content:    "From DB fallback",
		CreatedAt:  time.Now(),
		ReplyCount: 1,
		Upvotes:    2,
	}

	// No Redis insert

	comments, err := cache.ListComments(ctx, threadID, "upvotes", 999, 10, func(_ context.Context, _ uuid.UUID) ([]model.Comment, error) {
		return []model.Comment{c}, nil
	})
	require.NoError(t, err)
	require.Len(t, comments, 1)
	require.Equal(t, c.ID, comments[0].ID)
}

func TestUpdateCommentScore(t *testing.T) {
	ctx := context.Background()
	cache := setupRedis(t)

	threadID := uuid.New()
	c := model.Comment{
		ID:        uuid.New(),
		ThreadID:  threadID,
		UserID:    "user123",
		Content:   "Scored comment",
		CreatedAt: time.Now(),
		Upvotes:   5,
	}

	err := cache.SetComment(ctx, &c)
	require.NoError(t, err)

	err = cache.UpdateCommentScore(ctx, c.ID, "upvotes", 3)
	require.NoError(t, err)

	updated, err := cache.GetCommentByID(ctx, c.ID)
	require.NoError(t, err)
	require.Equal(t, 8, updated.Upvotes)
}

func TestSetComment_EvictsLowScore(t *testing.T) {
	ctx := context.Background()
	cache := setupRedis(t)
	threadID := uuid.New()

	// Insert 15 comments with increasing upvotes (from 1 to 15)
	for i := 1; i <= 15; i++ {
		c := &model.Comment{
			ID:         uuid.New(),
			ThreadID:   threadID,
			UserID:     fmt.Sprintf("user%d", i),
			Content:    fmt.Sprintf("Comment #%d", i),
			CreatedAt:  time.Now().Add(time.Duration(i) * time.Second),
			Upvotes:    i,
			ReplyCount: 0,
		}
		err := cache.SetComment(ctx, c)
		require.NoError(t, err)
	}

	// Redis should contain only top 10 comments by upvotes (6 to 15)
	zsetKey := fmt.Sprintf("comments:%s:upvotes", threadID)
	members, err := cache.client.ZRangeWithScores(ctx, zsetKey, 0, -1).Result()
	require.NoError(t, err)
	require.Len(t, members, 10)

	// First (lowest) score in top 10 should be 6
	require.Equal(t, float64(6), members[0].Score)
	require.Equal(t, float64(15), members[9].Score)
}
