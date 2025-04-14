package service

import (
	"context"
	"testing"

	"github.com/kiremitrov123/onboarding/texteditor/model"
	"github.com/kiremitrov123/onboarding/texteditor/service/mocks"
)

func TestApplyDelta_ReconstructDocument(t *testing.T) {
	deltas := []model.Edit{
		{User: "alice", Op: "insert", Index: 0, Text: "Hello "},
		{User: "bob", Op: "insert", Index: 6, Text: "world!"},
		{User: "charlie", Op: "delete", Index: 6, Text: "world!"},
		{User: "dana", Op: "insert", Index: 6, Text: "world!"},
	}

	doc := ""
	for _, edit := range deltas {
		doc = applyDelta(doc, edit)
	}

	expected := "Hello world!"
	if doc != expected {
		t.Errorf("Expected document to be %q, got %q", expected, doc)
	}
}

func TestApplyEdit_UpdatesStateAndPublishes(t *testing.T) {
	mock := &mocks.PubSubMock{
		PublishEditFunc: func(ctx context.Context, docID string, edit model.Edit) error {
			return nil
		},
	}

	svc := NewEditorService(mock)
	docID := "doc1"
	edit := model.Edit{
		User:  "alice",
		Op:    "insert",
		Index: 0,
		Text:  "Hello ",
	}

	result, err := svc.ApplyEdit(context.Background(), docID, edit)
	if err != nil {
		t.Fatalf("ApplyEdit failed: %v", err)
	}

	want := "Hello "
	if result != want {
		t.Errorf("Expected text %q, got %q", want, result)
	}

	// Ensure PublishEdit was called
	calls := mock.PublishEditCalls()
	if len(calls) != 1 {
		t.Fatalf("Expected 1 call to PublishEdit, got %d", len(calls))
	}
	if calls[0].Edit != edit {
		t.Errorf("Published edit mismatch: got %+v, want %+v", calls[0].Edit, edit)
	}
}

func TestGetDocument_ReturnsCorrectState(t *testing.T) {
	mock := &mocks.PubSubMock{
		PublishEditFunc: func(ctx context.Context, docID string, edit model.Edit) error {
			return nil
		},
	}
	svc := NewEditorService(mock)
	docID := "doc2"

	_, _ = svc.ApplyEdit(context.Background(), docID, model.Edit{
		User:  "bob",
		Op:    "insert",
		Index: 0,
		Text:  "World",
	})

	got := svc.GetDocument(docID)
	want := "World"

	if got != want {
		t.Errorf("Expected document %q, got %q", want, got)
	}
}

func TestApplyEdit_Delete(t *testing.T) {
	mock := &mocks.PubSubMock{
		PublishEditFunc: func(ctx context.Context, docID string, edit model.Edit) error {
			return nil
		},
	}
	svc := NewEditorService(mock)
	docID := "doc3"

	_, _ = svc.ApplyEdit(context.Background(), docID, model.Edit{
		User:  "alice",
		Op:    "insert",
		Index: 0,
		Text:  "Hello world!",
	})

	_, _ = svc.ApplyEdit(context.Background(), docID, model.Edit{
		User:  "charlie",
		Op:    "delete",
		Index: 6,
		Text:  "world!",
	})

	got := svc.GetDocument(docID)
	want := "Hello "

	if got != want {
		t.Errorf("Expected document %q, got %q", want, got)
	}
}
