package api_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sony/gobreaker"
	"github.com/stretchr/testify/assert"

	"github.com/kiremitrov123/onboarding/src/ogpreview/api"
	"github.com/kiremitrov123/onboarding/src/ogpreview/api/mocks"
	"github.com/kiremitrov123/onboarding/src/ogpreview/model"
)

// successfulBreaker always returns a successful OGTags response
type successfulBreaker struct{}

func (sb *successfulBreaker) Execute(fn func() (interface{}, error)) (interface{}, error) {
	return model.OGTags{
		Title:       "Extracted Title",
		Description: "Extracted description",
		Image:       "http://example.com/extracted.jpg",
	}, nil
}

// failingBreaker always returns an error
type failingBreaker struct{}

func (f *failingBreaker) Execute(fn func() (interface{}, error)) (interface{}, error) {
	return nil, errors.New("extraction failed")
}

func TestAPI_CacheHit(t *testing.T) {
	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{Name: "test"})

	mockCache := &mocks.CacheMock{
		GetTagsFunc: func(url string) (*model.OGTags, error) {
			return &model.OGTags{
				Title:       "Example",
				Description: "Example description",
				Image:       "http://example.com/image.jpg",
			}, nil
		},
		SetTagsFunc: func(url string, tags model.OGTags) error {
			return nil
		},
	}

	api := &api.API{Cache: mockCache, CB: cb}
	req := httptest.NewRequest("GET", "/preview?url=http://example.com", nil)
	rr := httptest.NewRecorder()
	api.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var tags model.OGTags
	err := json.NewDecoder(rr.Body).Decode(&tags)
	assert.NoError(t, err)
	assert.Equal(t, "Example", tags.Title)
}

func TestAPI_CircuitBreakerError(t *testing.T) {
	mockCache := &mocks.CacheMock{
		GetTagsFunc: func(url string) (*model.OGTags, error) {
			return nil, errors.New("cache miss")
		},
		SetTagsFunc: func(url string, tags model.OGTags) error {
			return nil
		},
	}

	api := &api.API{Cache: mockCache, CB: &failingBreaker{}}
	req := httptest.NewRequest("GET", "/preview?url=http://example.com", nil)
	rr := httptest.NewRecorder()
	api.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestAPI_ExtractionSuccess(t *testing.T) {
	mockCache := &mocks.CacheMock{
		GetTagsFunc: func(url string) (*model.OGTags, error) {
			return nil, errors.New("cache miss")
		},
		SetTagsFunc: func(url string, tags model.OGTags) error {
			return nil
		},
	}

	api := &api.API{Cache: mockCache, CB: &successfulBreaker{}}
	req := httptest.NewRequest("GET", "/preview?url=http://example.com", nil)
	rr := httptest.NewRecorder()
	api.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var tags model.OGTags
	err := json.NewDecoder(rr.Body).Decode(&tags)
	assert.NoError(t, err)
	assert.Equal(t, "Extracted Title", tags.Title)
}

func TestAPI_MissingURL(t *testing.T) {
	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{Name: "test"})

	mockCache := &mocks.CacheMock{
		GetTagsFunc: func(url string) (*model.OGTags, error) {
			return nil, errors.New("should not be called")
		},
		SetTagsFunc: func(url string, tags model.OGTags) error {
			return nil
		},
	}

	api := &api.API{Cache: mockCache, CB: cb}
	req := httptest.NewRequest("GET", "/preview", nil)
	rr := httptest.NewRecorder()
	api.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}
