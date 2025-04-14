package api

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/kiremitrov123/onboarding/stockprice/cache"
	"github.com/kiremitrov123/onboarding/stockprice/metrics"
	"github.com/kiremitrov123/onboarding/stockprice/model"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Cache interface {
	Get(ctx context.Context, symbol string) (*model.Price, error)
	Set(ctx context.Context, symbol string, price model.Price) error
	SubscribeInvalidation(ctx context.Context, onInvalidate func(key string)) error
}

type API struct {
	Local *cache.LocalCache
	Redis Cache

	once sync.Once
	mux  *http.ServeMux
}

func (a *API) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.once.Do(a.setupRoutes)
	a.mux.ServeHTTP(w, r)
}

func (a *API) setupRoutes() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /price", a.recordMetrics("GET /price", a.handleGetPrice))
	mux.HandleFunc("POST /price", a.recordMetrics("POST /price", a.handleSetPrice))
	mux.Handle("/metrics", promhttp.Handler())
	a.mux = mux
}

func (a *API) handleGetPrice(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	const symbol = "AAPL"

	if price, found := a.Local.Get(symbol); found {
		metrics.CacheHits.WithLabelValues(symbol).Inc()
		a.respond(w, http.StatusOK, price)
		return
	}
	metrics.CacheMisses.WithLabelValues(symbol).Inc()

	// Set placeholder to avoid race condition
	a.Local.SetPlaceholder(symbol)

	price, err := a.Redis.Get(ctx, symbol)
	if err != nil {
		a.Local.Invalidate(symbol)
		a.respondError(w, http.StatusNotFound, "price not found")
		return
	}

	a.Local.Set(symbol, *price)
	a.respond(w, http.StatusOK, price)
}

func (a *API) handleSetPrice(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var input struct {
		Price float64 `json:"price"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		a.respondError(w, http.StatusBadRequest, "invalid input")
		return
	}

	price := model.Price{
		Symbol:    "AAPL",
		Price:     input.Price,
		Timestamp: time.Now(),
	}

	if err := a.Redis.Set(ctx, "AAPL", price); err != nil {
		a.respondError(w, http.StatusInternalServerError, "failed to set price")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *API) recordMetrics(path string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		metrics.RequestsTotal.WithLabelValues(path, r.Method).Inc()
		next(w, r)
		duration := time.Since(start).Seconds()
		metrics.RequestDuration.WithLabelValues(path).Observe(duration)
	}
}

func (a *API) respond(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func (a *API) respondError(w http.ResponseWriter, status int, message string) {
	type response struct {
		Error string `json:"error"`
	}
	a.respond(w, status, response{Error: message})
}
