package model

import (
	"time"

	"github.com/google/uuid"
)

type Reaction struct {
	ID        uuid.UUID `json:"id"`
	CommentID uuid.UUID `json:"comment_id"`
	UserID    string    `json:"user_id"`
	Type      string    `json:"type"` // "like", "upvote" or "downvote"
	CreatedAt time.Time `json:"created_at"`
}
