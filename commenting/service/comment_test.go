package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/kiremitrov123/onboarding/commenting/model"
	"github.com/kiremitrov123/onboarding/commenting/service"
	"github.com/kiremitrov123/onboarding/commenting/service/mocks"
	"github.com/stretchr/testify/require"
)

func TestToggleReaction_InsertSuccess(t *testing.T) {
	ctx := context.Background()
	commentID := uuid.New()
	userID := "user1"
	reactionType := "like"
	field := "likes"

	repo := &mocks.CommentRepoMock{}
	cache := &mocks.CommentCacheMock{}
	svc := service.NewCommentService(repo, cache)

	repo.AddReactionFunc = func(ctx context.Context, r *model.Reaction) (bool, error) {
		require.Equal(t, commentID, r.CommentID)
		require.Equal(t, userID, r.UserID)
		require.Equal(t, reactionType, r.Type)
		return true, nil
	}

	repo.IncrementReactionCountFunc = func(ctx context.Context, id uuid.UUID, f string) error {
		require.Equal(t, commentID, id)
		require.Equal(t, field, f)
		return nil
	}

	cache.UpdateCommentScoreFunc = func(ctx context.Context, id uuid.UUID, f string, delta int) error {
		require.Equal(t, commentID, id)
		require.Equal(t, field, f)
		require.Equal(t, 1, delta)
		return nil
	}

	err := svc.ToggleReaction(ctx, commentID, userID, reactionType, field)
	require.NoError(t, err)
}

func TestToggleReaction_ToggleOff(t *testing.T) {
	ctx := context.Background()
	commentID := uuid.New()
	userID := "user2"
	reactionType := "upvote"
	field := "upvotes"

	repo := &mocks.CommentRepoMock{}
	cache := &mocks.CommentCacheMock{}
	svc := service.NewCommentService(repo, cache)

	repo.AddReactionFunc = func(ctx context.Context, r *model.Reaction) (bool, error) {
		return false, nil // simulate already exists
	}

	repo.DeleteReactionFunc = func(ctx context.Context, id uuid.UUID, user string, typ string) error {
		require.Equal(t, commentID, id)
		require.Equal(t, userID, user)
		require.Equal(t, reactionType, typ)
		return nil
	}

	repo.DecrementReactionCountFunc = func(ctx context.Context, id uuid.UUID, f string) error {
		require.Equal(t, commentID, id)
		require.Equal(t, field, f)
		return nil
	}

	cache.UpdateCommentScoreFunc = func(ctx context.Context, id uuid.UUID, f string, delta int) error {
		require.Equal(t, commentID, id)
		require.Equal(t, field, f)
		require.Equal(t, -1, delta)
		return nil
	}

	err := svc.ToggleReaction(ctx, commentID, userID, reactionType, field)
	require.NoError(t, err)
}

func TestToggleReaction_DBInsertError(t *testing.T) {
	ctx := context.Background()
	commentID := uuid.New()
	userID := "user3"
	reactionType := "downvote"
	field := "downvotes"

	repo := &mocks.CommentRepoMock{}
	cache := &mocks.CommentCacheMock{}
	svc := service.NewCommentService(repo, cache)

	repo.AddReactionFunc = func(ctx context.Context, r *model.Reaction) (bool, error) {
		return false, errors.New("db error")
	}

	err := svc.ToggleReaction(ctx, commentID, userID, reactionType, field)
	require.Error(t, err)
	require.Contains(t, err.Error(), "db error")
}
