package db

import (
	"time"

	"github.com/google/uuid"
	"github.com/kiremitrov123/onboarding/commenting/model"
	"github.com/uptrace/bun"
)

type CommentEntity struct {
	bun.BaseModel `bun:"table:comments"`

	ID         uuid.UUID  `bun:",pk,type:uuid,default:gen_random_uuid()"`
	ParentID   *uuid.UUID `bun:",nullzero"`
	ThreadID   uuid.UUID  `bun:",notnull"`
	UserID     string     `bun:",notnull"`
	Content    string     `bun:",notnull"`
	ReplyCount int        `bun:",notnull,default:0"`
	Upvotes    int        `bun:",notnull,default:0"`
	Downvotes  int        `bun:",notnull,default:0"`
	Likes      int        `bun:",notnull,default:0"`
	CreatedAt  time.Time  `bun:",nullzero,default::now()"`
}

type ReactionEntity struct {
	bun.BaseModel `bun:"table:comment_reactions"`

	ID        uuid.UUID `bun:",pk,type:uuid,default:gen_random_uuid()"`
	CommentID uuid.UUID `bun:",notnull"`
	UserID    string    `bun:",notnull"`
	Type      string    `bun:",notnull"`
	CreatedAt time.Time `bun:",nullzero,default::now()"`
}

func (c CommentEntity) APIComment() model.Comment {
	return model.Comment{
		ID:         c.ID,
		ParentID:   c.ParentID,
		ThreadID:   c.ThreadID,
		UserID:     c.UserID,
		Content:    c.Content,
		ReplyCount: c.ReplyCount,
		Upvotes:    c.Upvotes,
		Downvotes:  c.Downvotes,
		Likes:      c.Likes,
		CreatedAt:  c.CreatedAt,
	}
}

func (r ReactionEntity) APIReaction() model.Reaction {
	return model.Reaction{
		ID:        r.ID,
		CommentID: r.CommentID,
		UserID:    r.UserID,
		Type:      r.Type,
		CreatedAt: r.CreatedAt,
	}
}
