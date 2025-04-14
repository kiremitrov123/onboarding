package model

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
)

type Comment struct {
	ID         uuid.UUID  `json:"id"`
	ParentID   *uuid.UUID `json:"parent_id,omitempty"`
	ThreadID   uuid.UUID  `json:"thread_id"`
	UserID     string     `json:"user_id"`
	Content    string     `json:"content"`
	ReplyCount int        `json:"reply_count"`
	Upvotes    int        `json:"upvotes"`
	Downvotes  int        `json:"downvotes"`
	Likes      int        `json:"likes"`
	CreatedAt  time.Time  `json:"created_at"`
}

type QueryCommentsFunc func(ctx context.Context, threadID uuid.UUID) ([]Comment, error)

func (c *Comment) ToHash() map[string]interface{} {
	return map[string]interface{}{
		"id":          c.ID.String(),
		"parent_id":   uuidOrNil(c.ParentID),
		"thread_id":   c.ThreadID.String(),
		"user_id":     c.UserID,
		"content":     c.Content,
		"reply_count": c.ReplyCount,
		"upvotes":     c.Upvotes,
		"downvotes":   c.Downvotes,
		"likes":       c.Likes,
		"created_at":  c.CreatedAt.UnixNano(),
	}
}

func CommentFromHash(data map[string]string) (Comment, error) {
	id, err := uuid.Parse(data["id"])
	if err != nil {
		return Comment{}, fmt.Errorf("invalid id: %w", err)
	}

	threadID, err := uuid.Parse(data["thread_id"])
	if err != nil {
		return Comment{}, fmt.Errorf("invalid thread_id: %w", err)
	}

	var parentID *uuid.UUID
	if val := data["parent_id"]; val != "" && val != uuid.Nil.String() {
		parsed, err := uuid.Parse(val)
		if err == nil {
			parentID = &parsed
		}
	}

	createdAt, err := parseUnixNano(data["created_at"])
	if err != nil {
		return Comment{}, fmt.Errorf("invalid created_at: %w", err)
	}

	return Comment{
		ID:         id,
		ParentID:   parentID,
		ThreadID:   threadID,
		UserID:     data["user_id"],
		Content:    data["content"],
		ReplyCount: intFromStr(data["reply_count"]),
		Upvotes:    intFromStr(data["upvotes"]),
		Downvotes:  intFromStr(data["downvotes"]),
		Likes:      intFromStr(data["likes"]),
		CreatedAt:  createdAt,
	}, nil
}

func intFromStr(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}

func parseUnixNano(s string) (time.Time, error) {
	nanos, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(0, nanos), nil
}

func uuidOrNil(u *uuid.UUID) string {
	if u != nil {
		return u.String()
	}
	return uuid.Nil.String()
}
