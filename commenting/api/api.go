package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"sync"

	"github.com/google/uuid"
	"github.com/kiremitrov123/onboarding/commenting/model"
	"github.com/kiremitrov123/onboarding/commenting/service"
)

type API struct {
	Svc    *service.CommentService
	Logger *slog.Logger

	once sync.Once
	mux  *http.ServeMux
}

func NewAPI(svc *service.CommentService, logger *slog.Logger) *API {
	return &API{
		Svc:    svc,
		Logger: logger,
	}
}

func (a *API) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.once.Do(a.setupRoutes)
	a.mux.ServeHTTP(w, r)
}

func (a *API) setupRoutes() {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /comments", a.handleCreateComment)
	mux.HandleFunc("GET /comments", a.handleListComments)

	// TODO: Implement comment update (PATCH /comments/{id})
	// mux.HandleFunc("PATCH /comments/{id}", a.handleUpdateComment)
	// TODO: Implement comment delete (DELETE /comments/{id})
	// mux.HandleFunc("DELETE /comments/{id}", a.handleDeleteComment)

	mux.HandleFunc("POST /comments/{id}/upvote", a.handleReaction(a.Svc.Upvote))
	mux.HandleFunc("POST /comments/{id}/downvote", a.handleReaction(a.Svc.Downvote))
	mux.HandleFunc("POST /comments/{id}/like", a.handleReaction(a.Svc.Like))

	a.mux = mux
}

func (a *API) handleCreateComment(w http.ResponseWriter, r *http.Request) {
	var c model.Comment
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		a.Logger.Error("invalid input", slog.String("error", err.Error()))
		a.respondError(w, http.StatusBadRequest, "invalid input")
		return
	}

	if err := a.Svc.CreateComment(r.Context(), &c); err != nil {
		a.Logger.Error("failed to create comment",
			slog.String("user_id", c.UserID),
			slog.String("content", c.Content),
			slog.String("error", err.Error()),
		)
		a.respondError(w, http.StatusInternalServerError, "failed to create comment")
		return
	}

	a.Logger.Info("comment created",
		slog.String("id", c.ID.String()),
		slog.String("thread_id", c.ThreadID.String()),
		slog.String("user_id", c.UserID),
		slog.Time("created_at", c.CreatedAt),
	)

	w.Header().Set("Location", fmt.Sprintf("/comments/%s", c.ID))
	a.respond(w, http.StatusCreated, c)
}

func (a *API) handleListComments(w http.ResponseWriter, r *http.Request) {
	threadID := r.URL.Query().Get("thread_id")
	if threadID == "" {
		a.Logger.Warn("missing thread_id in query")
		a.respondError(w, http.StatusBadRequest, "missing thread_id")
		return
	}
	tid, err := uuid.Parse(threadID)
	if err != nil {
		a.Logger.Warn("invalid thread_id", slog.String("thread_id", threadID))
		a.respondError(w, http.StatusBadRequest, "invalid thread_id format")
		return
	}

	sort := r.URL.Query().Get("sort")

	limit := 10
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	cursorStr := r.URL.Query().Get("cursor")
	cursor, err := strconv.ParseInt(cursorStr, 10, 64)
	if cursorStr != "" && err != nil {
		a.Logger.Warn("invalid cursor", slog.String("cursor", cursorStr))
		a.respondError(w, http.StatusBadRequest, "invalid cursor value")
		return
	}

	var comments []model.Comment

	switch sort {
	case "upvotes":
		comments, err = a.Svc.ListByUpvotes(r.Context(), tid, cursor, limit)
	case "replies":
		comments, err = a.Svc.ListByReplies(r.Context(), tid, cursor, limit)
	default:
		comments, err = a.Svc.ListByDate(r.Context(), tid, cursor, limit)
	}

	if err != nil {
		a.Logger.Error("failed to list comments",
			slog.String("thread_id", tid.String()),
			slog.String("sort", sort),
			slog.String("error", err.Error()),
		)
		a.respondError(w, http.StatusInternalServerError, "failed to list comments")
		return
	}

	a.Logger.Info("listed comments",
		slog.String("thread_id", tid.String()),
		slog.String("sort", sort),
		slog.Int("count", len(comments)),
	)

	// Include next cursor in response
	var nextCursor int64
	if len(comments) > 0 {
		last := comments[len(comments)-1]
		switch sort {
		case "upvotes":
			nextCursor = int64(last.Upvotes)
		case "replies":
			nextCursor = int64(last.ReplyCount)
		default:
			nextCursor = last.CreatedAt.UnixNano()
		}
	}

	a.respond(w, http.StatusOK, map[string]interface{}{
		"comments":    comments,
		"next_cursor": nextCursor,
	})
}

type ReactionRequest struct {
	UserID string `json:"user_id"`
}

func (a *API) handleReaction(action func(context.Context, uuid.UUID, string) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		commentID, err := uuid.Parse(idStr)
		if err != nil {
			a.Logger.Warn("invalid comment ID", slog.String("id", idStr))
			a.respondError(w, http.StatusBadRequest, "invalid UUID format")
			return
		}

		var body ReactionRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.UserID == "" {
			a.Logger.Warn("invalid reaction payload")
			a.respondError(w, http.StatusBadRequest, "missing or invalid thread_id/user_id")
			return
		}

		if err := action(r.Context(), commentID, body.UserID); err != nil {
			a.Logger.Error("reaction failed",
				slog.String("comment_id", commentID.String()),
				slog.String("user_id", body.UserID),
				slog.String("type", r.URL.Path),
				slog.String("error", err.Error()),
			)
			a.respondError(w, http.StatusInternalServerError, "action failed")
			return
		}

		a.Logger.Info("reaction added",
			slog.String("comment_id", commentID.String()),
			slog.String("user_id", body.UserID),
			slog.String("action", r.URL.Path),
		)

		w.WriteHeader(http.StatusNoContent)
	}
}

func (a *API) respond(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func (a *API) respondError(w http.ResponseWriter, status int, message string) {
	type response struct {
		Error string `json:"error"`
	}
	a.respond(w, status, response{Error: message})
}
