package service

import (
	"context"
	"sync"

	"github.com/kiremitrov123/onboarding/texteditor/model"
	"github.com/kiremitrov123/onboarding/texteditor/redis"
)

type EditorService struct {
	pubsub redis.PubSub

	mu    sync.RWMutex
	state map[string]string
}

func NewEditorService(pubsub redis.PubSub) *EditorService {
	return &EditorService{
		pubsub: pubsub,
		state:  make(map[string]string),
	}
}

// Applies a delta to the current document and publishes it
func (s *EditorService) ApplyEdit(ctx context.Context, docID string, edit model.Edit) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	text := s.state[docID]
	newText := applyDelta(text, edit)
	s.state[docID] = newText

	if err := s.pubsub.PublishEdit(ctx, docID, edit); err != nil {
		return "", err
	}
	return newText, nil
}

// Returns the current text of a document
func (s *EditorService) GetDocument(docID string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state[docID]
}

// Subscribes to Redis and forwards incoming edits to the provided channel until the context is canceled.
func (s *EditorService) StreamEdits(ctx context.Context, docID string, ch chan<- model.Edit) error {
	return s.pubsub.SubscribeEdits(ctx, docID, func(edit model.Edit) {
		select {
		case ch <- edit:
		case <-ctx.Done():
		}
	})
}

// applyDelta applies an insert or delete operation to the text
func applyDelta(text string, edit model.Edit) string {
	switch edit.Op {
	case "insert":
		if edit.Index > len(text) {
			edit.Index = len(text)
		}
		return text[:edit.Index] + edit.Text + text[edit.Index:]

	case "delete":
		if edit.Index >= len(text) {
			return text
		}
		end := edit.Index + len(edit.Text)
		if end > len(text) {
			end = len(text)
		}
		return text[:edit.Index] + text[end:]

	default:
		return text
	}
}
