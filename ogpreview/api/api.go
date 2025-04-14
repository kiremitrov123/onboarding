package api

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/kiremitrov123/onboarding/src/ogpreview/model"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

type Cache interface {
	GetTags(url string) (*model.OGTags, error)
	SetTags(url string, tags model.OGTags) error
}

type CBreaker interface {
	Execute(func() (interface{}, error)) (interface{}, error)
}

type API struct {
	Cache Cache
	CB    CBreaker

	once sync.Once
	mux  *http.ServeMux
}

func (a *API) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.once.Do(a.setupRoutes)
	a.mux.ServeHTTP(w, r)
}

func (a *API) setupRoutes() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /preview", a.previewHandler)
	a.mux = mux
}

func (a *API) previewHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tr := otel.Tracer("og-preview-api")
	ctx, span := tr.Start(ctx, "PreviewHandler")
	defer span.End()

	url := r.URL.Query().Get("url")
	if url == "" {
		a.respondError(w, http.StatusBadRequest, "url param is required")
		return
	}
	span.SetAttributes(attribute.String("url", url))

	if tags, err := a.Cache.GetTags(url); err == nil && tags != nil {
		span.AddEvent("cache_hit")
		a.respond(w, http.StatusOK, tags)
		return
	}
	span.AddEvent("cache_miss")

	result, err := a.CB.Execute(func() (interface{}, error) {
		return FetchOGTags(ctx, url)
	})
	if err != nil {
		a.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	tags := result.(model.OGTags)
	a.Cache.SetTags(url, tags)
	a.respond(w, http.StatusOK, tags)
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
