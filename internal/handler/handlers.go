// Package handler provides HTTP request handlers for the message board API.
package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"message-board/internal/helper"
	"message-board/internal/response"
	"message-board/internal/store"
)

// Handlers holds HTTP handlers with access to the message store and logger.
type Handlers struct {
	store *store.Store
	log   *slog.Logger
}

// NewHandlers creates a new Handlers instance with slog.Default().
func NewHandlers(store *store.Store) *Handlers {
	return &Handlers{
		store: store,
		log:   slog.Default(),
	}
}

// HealthHandler handles GET /health — returns "ok".
func (h *Handlers) HealthHandler(w http.ResponseWriter, r *http.Request) {
	h.log.Debug("health check", "method", r.Method, "path", r.URL.Path)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

// EchoHandler handles POST /echo — echoes body back, with JSON support.
func (h *Handlers) EchoHandler(w http.ResponseWriter, r *http.Request) {
	body, err := helper.ReadBody(r)
	if err != nil {
		h.log.Error("failed to read request body", "path", r.URL.Path, "err", err)
		response.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	contentType := r.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "application/json") {
		h.echoJSON(w, body)
		return
	}

	h.log.Debug("echo text", "path", r.URL.Path, "size", len(body))
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}

// echoJSON parses {"message": "..."} and returns it as JSON.
func (h *Handlers) echoJSON(w http.ResponseWriter, body []byte) {
	var req struct {
		Message string `json:"message"`
	}
	if err := helper.JSONDecode(body, &req); err != nil {
		h.log.Error("invalid JSON in echo", "err", err)
		response.WriteError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	h.log.Debug("echo json", "message", req.Message)
	response.WriteJSON(w, http.StatusOK, map[string]string{"message": req.Message})
}

// MessagesHandler handles GET/POST /messages.
func (h *Handlers) MessagesHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.ListMessages(w, r)
	case http.MethodPost:
		h.CreateMessage(w, r)
	default:
		response.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// ListMessages handles GET /messages — returns all messages, newest first.
func (h *Handlers) ListMessages(w http.ResponseWriter, r *http.Request) {
	messages := h.store.List()
	if messages == nil {
		messages = []store.Message{}
	}
	h.log.Debug("list messages", "count", len(messages))
	response.WriteJSON(w, http.StatusOK, messages)
}

// CreateMessage handles POST /messages — creates a new message.
func (h *Handlers) CreateMessage(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Message string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Error("invalid JSON in create message", "err", err)
		response.WriteError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	defer r.Body.Close()

	if req.Message == "" {
		h.log.Error("empty message", "path", r.URL.Path)
		response.WriteError(w, http.StatusBadRequest, "message is required")
		return
	}

	msg := h.store.Create(req.Message)
	h.log.Debug("created message", "id", msg.ID, "message", msg.Message)
	response.WriteJSON(w, http.StatusCreated, msg)
}

// MessageByIDHandler handles DELETE /messages/{id}.
func (h *Handlers) MessageByIDHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if idStr == "" {
		if v := r.Context().Value("route:id"); v != nil {
			idStr = v.(string)
		}
	}
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		h.log.Error("invalid message ID", "id_str", idStr)
		response.WriteError(w, http.StatusBadRequest, "invalid message ID")
		return
	}

	if !h.store.Delete(id) {
		h.log.Error("message not found", "id", id)
		response.WriteError(w, http.StatusNotFound, "message not found")
		return
	}

	h.log.Debug("deleted message", "id", id)
	w.WriteHeader(http.StatusNoContent)
}
