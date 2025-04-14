package api

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/kiremitrov123/onboarding/texteditor/model"
	"github.com/kiremitrov123/onboarding/texteditor/service"
)

type API struct {
	Svc  *service.EditorService
	once sync.Once
	mux  *http.ServeMux
}

func NewAPI(svc *service.EditorService) *API {
	return &API{Svc: svc}
}

func (a *API) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.once.Do(a.setupRoutes)
	a.mux.ServeHTTP(w, r)
}

func (a *API) setupRoutes() {
	mux := http.NewServeMux()
	mux.HandleFunc("/subscribe", a.handleSubscribe)
	mux.HandleFunc("/edit", a.handleEdit)
	mux.HandleFunc("/document", a.handleGetDocument)
	a.mux = mux
}

func (a *API) handleEdit(w http.ResponseWriter, r *http.Request) {
	var req struct {
		DocID string     `json:"doc_id"`
		Edit  model.Edit `json:"edit"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println("Invalid input in /edit:", err)
		a.respondError(w, http.StatusBadRequest, "invalid input")
		return
	}

	log.Printf("Received edit: docID=%s, user=%s, op=%s, index=%d, text=%q",
		req.DocID, req.Edit.User, req.Edit.Op, req.Edit.Index, req.Edit.Text)

	_, err := a.Svc.ApplyEdit(r.Context(), req.DocID, req.Edit)
	if err != nil {
		log.Println("Failed to apply edit:", err)
		a.respondError(w, http.StatusInternalServerError, "failed to apply edit")
		return
	}

	a.respond(w, http.StatusOK, SuccessResponse{Status: "ok"})
}

func (a *API) handleGetDocument(w http.ResponseWriter, r *http.Request) {
	docID := r.URL.Query().Get("doc_id")
	if docID == "" {
		log.Println("Missing doc_id in /document")
		a.respondError(w, http.StatusBadRequest, "missing doc_id")
		return
	}

	log.Printf("Fetching document: docID=%s", docID)
	text := a.Svc.GetDocument(docID)
	a.respond(w, http.StatusOK, map[string]string{"text": text})
}

func (a *API) handleSubscribe(w http.ResponseWriter, r *http.Request) {
	docID := r.URL.Query().Get("doc_id")
	if docID == "" {
		log.Println("Missing doc_id in /subscribe")
		a.respondError(w, http.StatusBadRequest, "missing doc_id")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		log.Println("Streaming not supported by client")
		a.respondError(w, http.StatusInternalServerError, "streaming unsupported")
		return
	}

	editCh := make(chan model.Edit, 10)

	if err := a.Svc.StreamEdits(r.Context(), docID, editCh); err != nil {
		log.Println("Failed to subscribe to Redis PubSub:", err)
		a.respondError(w, http.StatusInternalServerError, "failed to subscribe to edits")
		return
	}

	_, _ = w.Write([]byte("data: connected\n\n"))
	flusher.Flush()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case edit := <-editCh:
			log.Printf("Streaming edit to client: docID=%s, user=%s, op=%s, index=%d, text=%q",
				docID, edit.User, edit.Op, edit.Index, edit.Text)
			if err := writeSSE(w, edit); err == nil {
				flusher.Flush()
			}
		case <-ticker.C:
			// Periodic heartbeat
			_, _ = w.Write([]byte("data: heartbeat\n\n"))
			flusher.Flush()
		case <-r.Context().Done():
			log.Printf("SSE connection closed for docID=%s", docID)
			return
		}
	}
}

func writeSSE(w http.ResponseWriter, data any) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	_, err = w.Write([]byte("data: " + string(bytes) + "\n\n"))
	return err
}

func (a *API) respond(w http.ResponseWriter, status int, body interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func (a *API) respondError(w http.ResponseWriter, status int, message string) {
	a.respond(w, status, ErrorResponse{Error: message})
}

type SuccessResponse struct {
	Status string `json:"status"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
