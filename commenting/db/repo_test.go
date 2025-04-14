package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kiremitrov123/onboarding/commenting/model"
	"github.com/stretchr/testify/require"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

var testRepo *Repo

func TestMain(m *testing.M) {
	dsn := "postgresql://root@localhost:26257/commenting?sslmode=disable"
	pg := pgdriver.NewConnector(pgdriver.WithDSN(dsn))
	sqlDB := bun.NewDB(sql.OpenDB(pg), pgdialect.New())

	testRepo = NewRepo(sqlDB)

	code := m.Run()
	os.Exit(code)
}

func insertTestComment(t *testing.T, threadID uuid.UUID, upvotes int) model.Comment {
	comment := model.Comment{
		ID:        uuid.New(),
		ThreadID:  threadID,
		UserID:    "test-user",
		Content:   fmt.Sprintf("Comment with upvotes %d", upvotes),
		Upvotes:   upvotes,
		CreatedAt: time.Now(),
	}
	err := testRepo.CreateComment(context.Background(), &comment)
	require.NoError(t, err)
	return comment
}

func TestListCommentsSorted(t *testing.T) {
	ctx := context.Background()
	threadID := uuid.New()

	// Insert test comments
	c1 := insertTestComment(t, threadID, 10)
	c2 := insertTestComment(t, threadID, 7)
	c3 := insertTestComment(t, threadID, 3)

	// Page 1: should get c1 and c2
	comments, err := testRepo.ListCommentsSorted(ctx, threadID, "upvotes", 11, 2)
	require.NoError(t, err)
	require.Len(t, comments, 2)
	require.Equal(t, c1.ID, comments[0].ID)
	require.Equal(t, c2.ID, comments[1].ID)

	// Page 2: should get c3
	comments, err = testRepo.ListCommentsSorted(ctx, threadID, "upvotes", 7, 10)
	require.NoError(t, err)
	require.Len(t, comments, 1)
	require.Equal(t, c3.ID, comments[0].ID)
}

func TestListCommentsSorted_InvalidSortField(t *testing.T) {
	ctx := context.Background()
	threadID := uuid.New()

	_, err := testRepo.ListCommentsSorted(ctx, threadID, "notarealfield", 0, 5)
	require.Error(t, err)
}

func TestListCommentsSorted_EmptyThread(t *testing.T) {
	ctx := context.Background()
	threadID := uuid.New()

	comments, err := testRepo.ListCommentsSorted(ctx, threadID, "upvotes", 9999, 5)
	require.NoError(t, err)
	require.Len(t, comments, 0)
}

func TestListCommentsSorted_LimitZero(t *testing.T) {
	ctx := context.Background()
	threadID := uuid.New()
	insertTestComment(t, threadID, 10)

	comments, err := testRepo.ListCommentsSorted(ctx, threadID, "upvotes", 9999, 0)
	require.NoError(t, err)
	require.Len(t, comments, 0)
}

func TestListCommentsSorted_CursorZero(t *testing.T) {
	ctx := context.Background()
	threadID := uuid.New()
	c1 := insertTestComment(t, threadID, 99)
	c2 := insertTestComment(t, threadID, 88)

	comments, err := testRepo.ListCommentsSorted(ctx, threadID, "upvotes", 0, 5)
	require.NoError(t, err)
	require.Len(t, comments, 2)
	require.Equal(t, c1.ID, comments[0].ID)
	require.Equal(t, c2.ID, comments[1].ID)
}

func TestListCommentsSorted_LimitGreaterThanTotal(t *testing.T) {
	ctx := context.Background()
	threadID := uuid.New()
	c1 := insertTestComment(t, threadID, 1)
	c2 := insertTestComment(t, threadID, 2)

	comments, err := testRepo.ListCommentsSorted(ctx, threadID, "upvotes", 9999, 10)
	require.NoError(t, err)
	require.Len(t, comments, 2)
	require.Equal(t, c2.ID, comments[0].ID)
	require.Equal(t, c1.ID, comments[1].ID)
}
