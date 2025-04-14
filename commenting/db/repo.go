package db

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kiremitrov123/onboarding/commenting/model"
	"github.com/uptrace/bun"
)

type Repo struct {
	DB *bun.DB
}

func NewRepo(db *bun.DB) *Repo {
	return &Repo{DB: db}
}

// CreateComment inserts a new comment into the database.
func (r *Repo) CreateComment(ctx context.Context, comment *model.Comment) error {
	entity := CommentEntity{
		ID:         comment.ID,
		ParentID:   comment.ParentID,
		ThreadID:   comment.ThreadID,
		UserID:     comment.UserID,
		Content:    comment.Content,
		ReplyCount: comment.ReplyCount,
		Upvotes:    comment.Upvotes,
		Downvotes:  comment.Downvotes,
		Likes:      comment.Likes,
		CreatedAt:  comment.CreatedAt,
	}
	_, err := r.DB.NewInsert().Model(&entity).Exec(ctx)
	return err
}

// GetCommentByID retrieves a comment record from the database.
func (r *Repo) GetCommentByID(ctx context.Context, commentID uuid.UUID) (*model.Comment, error) {
	var entity CommentEntity
	err := r.DB.NewSelect().
		Model(&entity).
		Where("id = ?", commentID).
		Limit(1).
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	return &model.Comment{
		ID:         entity.ID,
		ParentID:   entity.ParentID,
		ThreadID:   entity.ThreadID,
		UserID:     entity.UserID,
		Content:    entity.Content,
		ReplyCount: entity.ReplyCount,
		Upvotes:    entity.Upvotes,
		Downvotes:  entity.Downvotes,
		Likes:      entity.Likes,
		CreatedAt:  entity.CreatedAt,
	}, nil
}

// IncrementReplyCount increases the reply count by 1 for a parent comment.
func (r *Repo) IncrementReplyCount(ctx context.Context, parentID uuid.UUID) error {
	_, err := r.DB.NewUpdate().
		Model((*CommentEntity)(nil)).
		Where("id = ?", parentID).
		Set("reply_count = reply_count + 1").
		Exec(ctx)
	return err
}

// ListCommentsSorted fetches comments by thread ID sorted by the specified field.
func (r *Repo) ListCommentsSorted(ctx context.Context, threadID uuid.UUID, sortField string, cursor int64, limit int) ([]model.Comment, error) {
	if limit == 0 {
		return []model.Comment{}, nil
	}

	var entities []CommentEntity

	q := r.DB.NewSelect().
		Model(&entities).
		Where("thread_id = ?", threadID)

	// Pagination cursor
	if cursor > 0 {
		if sortField == "created_at" {
			t := time.Unix(0, cursor)
			q = q.Where("created_at < ?", t)
		} else {
			q = q.Where(fmt.Sprintf("%s < ?", sortField), cursor)
		}
	}

	// Order + Limit
	err := q.
		Order(fmt.Sprintf("%s DESC", sortField)).
		Limit(limit).
		Scan(ctx)

	if err != nil {
		return nil, err
	}

	out := make([]model.Comment, 0, len(entities))
	for _, e := range entities {
		out = append(out, e.APIComment())
	}
	return out, nil
}

// AddReaction stores a new reaction in the database.
func (r *Repo) AddReaction(ctx context.Context, reaction *model.Reaction) (bool, error) {
	entity := ReactionEntity{
		ID:        reaction.ID,
		CommentID: reaction.CommentID,
		UserID:    reaction.UserID,
		Type:      reaction.Type,
		CreatedAt: reaction.CreatedAt,
	}
	res, err := r.DB.NewInsert().
		Model(&entity).
		On("CONFLICT (comment_id, user_id, type) DO NOTHING").
		Exec(ctx)

	if err != nil {
		return false, err
	}

	rows, _ := res.RowsAffected()
	return rows > 0, nil
}

// DeleteReaction removes an existing reaction in the database.
func (r *Repo) DeleteReaction(ctx context.Context, commentID uuid.UUID, userID string, reactionType string) error {
	_, err := r.DB.NewDelete().
		Model((*ReactionEntity)(nil)).
		Where("comment_id = ?", commentID).
		Where("user_id = ?", userID).
		Where("type = ?", reactionType).
		Exec(ctx)
	return err
}

// IncrementReactionCount increments a specific counter field (e.g. likes, upvotes).
func (r *Repo) IncrementReactionCount(ctx context.Context, commentID uuid.UUID, field string) error {
	_, err := r.DB.NewUpdate().
		Model((*CommentEntity)(nil)).
		Where("id = ?", commentID).
		Set(fmt.Sprintf("%s = %s + 1", field, field)).
		Exec(ctx)
	return err
}

// DecrementReactionCount decrements a specific counter field (e.g. likes, upvotes).
func (r *Repo) DecrementReactionCount(ctx context.Context, commentID uuid.UUID, field string) error {
	_, err := r.DB.NewUpdate().
		Model((*CommentEntity)(nil)).
		Where("id = ?", commentID).
		Set(fmt.Sprintf("%s = %s - 1", field, field)).
		Exec(ctx)
	return err
}
