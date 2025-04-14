package redis

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/kiremitrov123/onboarding/commenting/model"
	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client *redis.Client
}

func NewCache(ctx context.Context, addr string) (*RedisCache, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
	})
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis at %s: %w", addr, err)
	}
	return &RedisCache{client: rdb}, nil
}

const (
	prefix   = "comments"
	ttl      = time.Hour
	maxItems = 10
)

// SetComment stores a comment as a hash and updates the sorted sets for date, replies, and upvotes.
func (rc *RedisCache) SetComment(ctx context.Context, c *model.Comment) error {
	threadID := c.ThreadID.String()
	commentKey := fmt.Sprintf("%s:%s", prefix, c.ID.String())
	data := c.ToHash()

	sortedScores := map[string]float64{
		"created_at": float64(c.CreatedAt.UnixNano()),
		"replies":    float64(c.ReplyCount),
		"upvotes":    float64(c.Upvotes),
	}

	return rc.client.Watch(ctx, func(tx *redis.Tx) error {
		_, err := tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			pipe.HSet(ctx, commentKey, data)
			pipe.Expire(ctx, commentKey, ttl)

			for field, score := range sortedScores {
				zKey := fmt.Sprintf("%s:%s:%s", prefix, threadID, field)
				pipe.ZAdd(ctx, zKey, redis.Z{Score: score, Member: commentKey})
				pipe.ZRemRangeByRank(ctx, zKey, 0, int64(-maxItems-1))
				pipe.Expire(ctx, zKey, ttl)
			}
			return nil
		})
		return err
	}, commentKey)
}

func (rc *RedisCache) GetCommentByID(ctx context.Context, commentID uuid.UUID) (*model.Comment, error) {
	commentKey := fmt.Sprintf("%s:%s", prefix, commentID.String())

	fields, err := rc.client.HGetAll(ctx, commentKey).Result()
	if err != nil {
		return nil, fmt.Errorf("redis get failed: %w", err)
	}
	if len(fields) == 0 {
		return nil, redis.Nil
	}

	comment, err := model.CommentFromHash(fields)
	if err != nil {
		return nil, fmt.Errorf("failed to parse comment from redis: %w", err)
	}

	return &comment, nil
}

// ListComments retrieves sorted comments from Redis or uses fallback to load and repopulate them.
func (rc *RedisCache) ListComments(
	ctx context.Context,
	threadID uuid.UUID,
	sortKey string,
	cursor int64,
	limit int,
	fallback model.QueryCommentsFunc,
) ([]model.Comment, error) {
	threadKey := threadID.String()
	zsetKey := fmt.Sprintf("%s:%s:%s", prefix, threadKey, sortKey)

	// Default to max value if cursor is not provided
	if cursor == 0 {
		cursor = math.MaxInt64
	}

	keys, err := rc.client.ZRevRangeByScore(ctx, zsetKey, &redis.ZRangeBy{
		Max:    fmt.Sprintf("(%d", cursor), // exclusive
		Min:    "-inf",
		Offset: 0,
		Count:  int64(limit),
	}).Result()

	if err != nil || len(keys) == 0 {
		comments, err := fallback(ctx, threadID)
		if err != nil {
			return nil, err
		}
		for _, c := range comments {
			_ = rc.SetComment(ctx, &c)
		}
		return comments, nil
	}

	var out []model.Comment
	for _, k := range keys {
		fields, err := rc.client.HGetAll(ctx, k).Result()
		if err != nil || len(fields) == 0 {
			continue
		}
		comment, err := model.CommentFromHash(fields)
		if err == nil {
			out = append(out, comment)
		}
	}
	return out, nil
}

// UpdateCommentScore increments a numeric field and updates the score in the sorted set.
func (rc *RedisCache) UpdateCommentScore(ctx context.Context, commentID uuid.UUID, field string, delta int) error {
	commentKey := fmt.Sprintf("%s:%s", prefix, commentID.String())

	return rc.client.Watch(ctx, func(tx *redis.Tx) error {
		fields, err := tx.HGetAll(ctx, commentKey).Result()
		if err != nil || len(fields) == 0 {
			return nil // silently ignore if not cached
		}

		threadID := fields["thread_id"]
		currentVal, _ := strconv.Atoi(fields[field])
		newVal := currentVal + delta

		zsetKey := fmt.Sprintf("%s:%s:%s", prefix, threadID, field)

		_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			pipe.HSet(ctx, commentKey, field, newVal)
			pipe.ZAdd(ctx, zsetKey, redis.Z{Score: float64(newVal), Member: commentKey})
			return nil
		})
		return err
	}, commentKey)
}
