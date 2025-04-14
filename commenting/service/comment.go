package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/kiremitrov123/onboarding/commenting/model"
)

type CommentRepo interface {
	CreateComment(ctx context.Context, comment *model.Comment) error
	GetCommentByID(ctx context.Context, commentID uuid.UUID) (*model.Comment, error)
	IncrementReplyCount(ctx context.Context, parentID uuid.UUID) error
	AddReaction(ctx context.Context, reaction *model.Reaction) (bool, error)
	DeleteReaction(ctx context.Context, commentID uuid.UUID, userID, reactionType string) error
	IncrementReactionCount(ctx context.Context, commentID uuid.UUID, field string) error
	DecrementReactionCount(ctx context.Context, commentID uuid.UUID, field string) error
	ListCommentsSorted(ctx context.Context, threadID uuid.UUID, sortField string, cursor int64, limit int) ([]model.Comment, error)
}

type CommentCache interface {
	SetComment(ctx context.Context, comment *model.Comment) error
	GetCommentByID(ctx context.Context, commentID uuid.UUID) (*model.Comment, error)
	UpdateCommentScore(ctx context.Context, commentID uuid.UUID, field string, delta int) error
	ListComments(ctx context.Context, threadID uuid.UUID, sortKey string, cursor int64, limit int, fallback model.QueryCommentsFunc) ([]model.Comment, error)
}

type CommentService struct {
	repo  CommentRepo
	cache CommentCache
}

func NewCommentService(repo CommentRepo, cache CommentCache) *CommentService {
	return &CommentService{repo: repo, cache: cache}
}

// CreateComment stores the comment in DB and cache, and updates parent reply count if needed.
func (s *CommentService) CreateComment(ctx context.Context, comment *model.Comment) error {
	// Generate a new UUID for the comment if none was provided
	if comment.ID == uuid.Nil {
		comment.ID = uuid.New()
	}

	// If it's a top-level comment, use its own ID as the thread ID
	// It's a reply — get parent comment to inherit thread ID
	if comment.ParentID == nil {
		comment.ThreadID = comment.ID
	} else {
		parent, err := s.GetCommentByID(ctx, *comment.ParentID)
		if err != nil {
			return err
		}
		comment.ThreadID = parent.ThreadID
	}

	if err := s.repo.CreateComment(ctx, comment); err != nil {
		return err
	}

	if comment.ParentID != nil {
		if err := s.repo.IncrementReplyCount(ctx, *comment.ParentID); err != nil {
			return err
		}
	}
	return s.cache.SetComment(ctx, comment)
}

// GetCommentByID retrieves the comment by its ID
// Tries cache, fallbacks to DB
func (s *CommentService) GetCommentByID(ctx context.Context, commentID uuid.UUID) (*model.Comment, error) {
	comment, err := s.cache.GetCommentByID(ctx, commentID)
	if err == nil {
		return comment, nil
	}

	comment, err = s.repo.GetCommentByID(ctx, commentID)
	if err != nil {
		return nil, err
	}

	return comment, nil
}

func (s *CommentService) Upvote(ctx context.Context, commentID uuid.UUID, userID string) error {
	return s.ToggleReaction(ctx, commentID, userID, "upvote", "upvotes")
}

func (s *CommentService) Downvote(ctx context.Context, commentID uuid.UUID, userID string) error {
	return s.ToggleReaction(ctx, commentID, userID, "downvote", "downvotes")
}

func (s *CommentService) Like(ctx context.Context, commentID uuid.UUID, userID string) error {
	return s.ToggleReaction(ctx, commentID, userID, "like", "likes")
}

func (s *CommentService) ListByDate(ctx context.Context, threadID uuid.UUID, cursor int64, limit int) ([]model.Comment, error) {
	return s.listSorted(ctx, threadID, "date", cursor, limit)
}

func (s *CommentService) ListByUpvotes(ctx context.Context, threadID uuid.UUID, cursor int64, limit int) ([]model.Comment, error) {
	return s.listSorted(ctx, threadID, "upvotes", cursor, limit)
}

func (s *CommentService) ListByReplies(ctx context.Context, threadID uuid.UUID, cursor int64, limit int) ([]model.Comment, error) {
	return s.listSorted(ctx, threadID, "replies", cursor, limit)
}

// ToggleReaction adds or removes a user reaction and adjusts the comment's score field to reflect the change.
func (s *CommentService) ToggleReaction(ctx context.Context, commentID uuid.UUID, userID, reactionType, field string) error {
	reaction := &model.Reaction{
		CommentID: commentID,
		UserID:    userID,
		Type:      reactionType,
	}

	toggledOn, err := s.repo.AddReaction(ctx, reaction)
	if err != nil {
		return err
	}

	if toggledOn {
		if err := s.repo.IncrementReactionCount(ctx, commentID, field); err != nil {
			return err
		}
		return s.cache.UpdateCommentScore(ctx, commentID, field, +1)
	}

	// toggled off
	if err := s.repo.DeleteReaction(ctx, commentID, userID, reactionType); err != nil {
		return err
	}
	if err := s.repo.DecrementReactionCount(ctx, commentID, field); err != nil {
		return err
	}
	return s.cache.UpdateCommentScore(ctx, commentID, field, -1)
}

// listSorted fetches from Redis or falls back to DB
// listing is based on the sort field
func (s *CommentService) listSorted(ctx context.Context, threadID uuid.UUID, sortField string, cursor int64, limit int) ([]model.Comment, error) {
	validFields := map[string]string{
		"date":    "created_at",
		"upvotes": "upvotes",
		"replies": "reply_count",
	}

	field, ok := validFields[sortField]
	if !ok {
		return nil, fmt.Errorf("invalid sort field: %s", sortField)
	}

	return s.cache.ListComments(ctx, threadID, field, cursor, limit, func(ctx context.Context, tid uuid.UUID) ([]model.Comment, error) {
		return s.repo.ListCommentsSorted(ctx, tid, field, cursor, limit)
	})
}
