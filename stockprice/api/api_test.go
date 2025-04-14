package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kiremitrov123/onboarding/stockprice/api"
	"github.com/kiremitrov123/onboarding/stockprice/api/mocks"
	"github.com/kiremitrov123/onboarding/stockprice/cache"
	"github.com/kiremitrov123/onboarding/stockprice/model"
)

func TestHandleGetPrice_CacheHit(t *testing.T) {
	mockLocal := cache.NewLocalCache(5*time.Minute, 10*time.Minute)
	mockRedis := &mocks.CacheMock{
		GetFunc: func(ctx context.Context, symbol string) (*model.Price, error) {
			return &model.Price{Symbol: "AAPL", Price: 123.45, Timestamp: time.Now()}, nil
		},
	}

	a := &api.API{
		Local: mockLocal,
		Redis: mockRedis,
	}

	req := httptest.NewRequest("GET", "/price", nil)
	rr := httptest.NewRecorder()
	a.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestHandleSetPrice_Success(t *testing.T) {
	mockRedis := &mocks.CacheMock{
		SetFunc: func(ctx context.Context, symbol string, price model.Price) error {
			return nil
		},
	}

	a := &api.API{
		Local: cache.NewLocalCache(time.Minute, time.Minute),
		Redis: mockRedis,
	}

	body := map[string]float64{"price": 456.78}
	data, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/price", bytes.NewReader(data))
	rr := httptest.NewRecorder()
	a.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rr.Code)
	}
}

func TestHandleSetPrice_InvalidInput(t *testing.T) {
	a := &api.API{}
	req := httptest.NewRequest("POST", "/price", bytes.NewReader([]byte(`{invalid json`)))
	rr := httptest.NewRecorder()
	a.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestHandleGetPrice_RedisMiss(t *testing.T) {
	mockLocal := cache.NewLocalCache(time.Minute, time.Minute)
	mockRedis := &mocks.CacheMock{
		GetFunc: func(ctx context.Context, symbol string) (*model.Price, error) {
			return nil, errors.New("not found")
		},
	}

	a := &api.API{
		Local: mockLocal,
		Redis: mockRedis,
	}

	req := httptest.NewRequest("GET", "/price", nil)
	rr := httptest.NewRecorder()
	a.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}
