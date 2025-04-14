package api

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kiremitrov123/onboarding/texteditor/model"
	"github.com/kiremitrov123/onboarding/texteditor/service"
	"github.com/kiremitrov123/onboarding/texteditor/service/mocks"
)

func newAPIWithMockedService() *API {
	mock := &mocks.PubSubMock{
		PublishEditFunc: func(ctx context.Context, docID string, edit model.Edit) error {
			return nil
		},
	}
	svc := service.NewEditorService(mock)
	return NewAPI(svc)
}

func TestHandleEdit_ValidInput(t *testing.T) {
	apiHandler := newAPIWithMockedService()

	reqBody := `{
		"doc_id": "doc1",
		"edit": {
			"user": "test",
			"op": "insert",
			"index": 0,
			"text": "Hello"
		}
	}`

	req := httptest.NewRequest(http.MethodPost, "/edit", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	apiHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", rr.Code)
	}
}

func TestHandleEdit_InvalidJSON(t *testing.T) {
	apiHandler := newAPIWithMockedService()

	reqBody := `{"invalid": true`

	req := httptest.NewRequest(http.MethodPost, "/edit", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	apiHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 Bad Request, got %d", rr.Code)
	}
}

func TestHandleGetDocument_Valid(t *testing.T) {
	mock := &mocks.PubSubMock{
		PublishEditFunc: func(ctx context.Context, docID string, edit model.Edit) error {
			return nil
		},
	}
	svc := service.NewEditorService(mock)

	// Preload state
	_, _ = svc.ApplyEdit(context.Background(), "doc2", model.Edit{
		User:  "alice",
		Op:    "insert",
		Index: 0,
		Text:  "Hello!",
	})

	apiHandler := NewAPI(svc)

	req := httptest.NewRequest(http.MethodGet, "/document?doc_id=doc2", nil)
	rr := httptest.NewRecorder()

	apiHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", rr.Code)
	}
}

func TestHandleGetDocument_MissingDocID(t *testing.T) {
	apiHandler := newAPIWithMockedService()

	req := httptest.NewRequest(http.MethodGet, "/document", nil)
	rr := httptest.NewRecorder()

	apiHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 Bad Request for missing doc_id, got %d", rr.Code)
	}
}
