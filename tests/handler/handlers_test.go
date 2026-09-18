package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"message-board/internal/handler"
	"message-board/internal/store"
)

func newTestHandlers(t *testing.T) *handler.Handlers {
	t.Helper()
	return handler.NewHandlers(store.NewStore())
}

func assertStatusCode(t *testing.T, w *httptest.ResponseRecorder, expected int) {
	t.Helper()
	if w.Code != expected {
		t.Errorf("expected status %d, got %d", expected, w.Code)
	}
}

func assertContentType(t *testing.T, w *httptest.ResponseRecorder, expected string) {
	t.Helper()
	ct := w.Header().Get("Content-Type")
	if ct != expected {
		t.Errorf("expected Content-Type %q, got %q", expected, ct)
	}
}

// --- Health ---

func TestHealthHandler_Success(t *testing.T) {
	h := newTestHandlers(t)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	h.HealthHandler(w, req)

	assertStatusCode(t, w, http.StatusOK)
	if !bytes.Equal(bytes.TrimSpace(w.Body.Bytes()), []byte("ok")) {
		t.Errorf("expected body 'ok', got %q", w.Body.String())
	}
}

func TestHealthHandler_WrongMethod(t *testing.T) {
	h := newTestHandlers(t)
	req := httptest.NewRequest(http.MethodPost, "/health", nil)
	w := httptest.NewRecorder()

	h.HealthHandler(w, req)

	assertStatusCode(t, w, http.StatusOK)
}

// --- Echo: plain text ---

func TestEchoHandler_PlainText(t *testing.T) {
	h := newTestHandlers(t)
	body := bytes.NewReader([]byte("hello world"))
	req := httptest.NewRequest(http.MethodPost, "/echo", body)
	req.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()

	h.EchoHandler(w, req)

	assertStatusCode(t, w, http.StatusOK)
	if !bytes.Equal(bytes.TrimSpace(w.Body.Bytes()), []byte("hello world")) {
		t.Errorf("expected 'hello world', got %q", w.Body.String())
	}
}

func TestEchoHandler_EmptyBody(t *testing.T) {
	h := newTestHandlers(t)
	req := httptest.NewRequest(http.MethodPost, "/echo", bytes.NewReader([]byte{}))
	w := httptest.NewRecorder()

	h.EchoHandler(w, req)

	assertStatusCode(t, w, http.StatusOK)
}

func TestEchoHandler_JSON(t *testing.T) {
	h := newTestHandlers(t)
	payload := `{"message": "hi"}`
	req := httptest.NewRequest(http.MethodPost, "/echo", bytes.NewReader([]byte(payload)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.EchoHandler(w, req)

	assertStatusCode(t, w, http.StatusOK)
	assertContentType(t, w, "application/json")

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["message"] != "hi" {
		t.Errorf("expected message 'hi', got %q", resp["message"])
	}
}

func TestEchoHandler_JSON_InvalidJSON(t *testing.T) {
	h := newTestHandlers(t)
	req := httptest.NewRequest(http.MethodPost, "/echo", bytes.NewReader([]byte("{invalid")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.EchoHandler(w, req)

	assertStatusCode(t, w, http.StatusBadRequest)
	assertContentType(t, w, "application/json")
}

func TestEchoHandler_JSON_Charset(t *testing.T) {
	h := newTestHandlers(t)
	payload := `{"message": "charset test"}`
	req := httptest.NewRequest(http.MethodPost, "/echo", bytes.NewReader([]byte(payload)))
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	w := httptest.NewRecorder()

	h.EchoHandler(w, req)

	assertStatusCode(t, w, http.StatusOK)
}

// --- Messages: Create ---

func TestCreateMessage_Success(t *testing.T) {
	h := newTestHandlers(t)
	payload := `{"message": "hello"}`
	req := httptest.NewRequest(http.MethodPost, "/messages", bytes.NewReader([]byte(payload)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.CreateMessage(w, req)

	assertStatusCode(t, w, http.StatusCreated)
	assertContentType(t, w, "application/json")

	var msg store.Message
	json.Unmarshal(w.Body.Bytes(), &msg)
	if msg.ID != 1 {
		t.Errorf("expected ID 1, got %d", msg.ID)
	}
	if msg.Message != "hello" {
		t.Errorf("expected message 'hello', got '%s'", msg.Message)
	}
	if msg.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
}

func TestCreateMessage_EmptyMessage(t *testing.T) {
	h := newTestHandlers(t)
	payload := `{"message": ""}`
	req := httptest.NewRequest(http.MethodPost, "/messages", bytes.NewReader([]byte(payload)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.CreateMessage(w, req)

	assertStatusCode(t, w, http.StatusBadRequest)
}

func TestCreateMessage_InvalidJSON(t *testing.T) {
	h := newTestHandlers(t)
	req := httptest.NewRequest(http.MethodPost, "/messages", bytes.NewReader([]byte("{bad")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.CreateMessage(w, req)

	assertStatusCode(t, w, http.StatusBadRequest)
}

func TestCreateMessage_MissingField(t *testing.T) {
	h := newTestHandlers(t)
	req := httptest.NewRequest(http.MethodPost, "/messages", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.CreateMessage(w, req)

	assertStatusCode(t, w, http.StatusBadRequest)
}

func TestMessagesHandler_ListOnGET(t *testing.T) {
	h := newTestHandlers(t)
	req := httptest.NewRequest(http.MethodGet, "/messages", nil)
	w := httptest.NewRecorder()

	h.MessagesHandler(w, req)

	assertStatusCode(t, w, http.StatusOK)
}

func TestMessagesHandler_MethodNotAllowed(t *testing.T) {
	h := newTestHandlers(t)
	req := httptest.NewRequest(http.MethodPut, "/messages", nil)
	w := httptest.NewRecorder()

	h.MessagesHandler(w, req)

	assertStatusCode(t, w, http.StatusMethodNotAllowed)
}

// --- Messages: List ---

func TestListMessages_Empty(t *testing.T) {
	h := newTestHandlers(t)
	req := httptest.NewRequest(http.MethodGet, "/messages", nil)
	w := httptest.NewRecorder()

	h.ListMessages(w, req)

	assertStatusCode(t, w, http.StatusOK)
	assertContentType(t, w, "application/json")

	var messages []store.Message
	json.Unmarshal(w.Body.Bytes(), &messages)
	if len(messages) != 0 {
		t.Errorf("expected empty array, got %v", messages)
	}
}

func TestListMessages_Order(t *testing.T) {
	s := store.NewStore()
	s.Create("first")
	s.Create("second")
	s.Create("third")
	h := handler.NewHandlers(s)

	req := httptest.NewRequest(http.MethodGet, "/messages", nil)
	w := httptest.NewRecorder()

	h.ListMessages(w, req)

	assertStatusCode(t, w, http.StatusOK)

	var messages []store.Message
	json.Unmarshal(w.Body.Bytes(), &messages)
	if len(messages) != 3 {
		t.Fatalf("expected 3 messages, got %d", len(messages))
	}
	if messages[0].Message != "third" {
		t.Errorf("expected first element 'third', got '%s'", messages[0].Message)
	}
}

// --- Messages: Delete ---

func TestDeleteMessage_Success(t *testing.T) {
	s := store.NewStore()
	s.Create("to delete")
	h := handler.NewHandlers(s)

	req := httptest.NewRequest(http.MethodDelete, "/messages/1", nil)
	ctx := routeContext(req, "id", "1")
	w := httptest.NewRecorder()

	h.MessageByIDHandler(w, req.WithContext(ctx))

	assertStatusCode(t, w, http.StatusNoContent)

	messages := s.List()
	if len(messages) != 0 {
		t.Errorf("expected 0 messages after delete, got %d", len(messages))
	}
}

func TestDeleteMessage_NotFound(t *testing.T) {
	h := newTestHandlers(t)

	req := httptest.NewRequest(http.MethodDelete, "/messages/999", nil)
	ctx := routeContext(req, "id", "999")
	w := httptest.NewRecorder()

	h.MessageByIDHandler(w, req.WithContext(ctx))

	assertStatusCode(t, w, http.StatusNotFound)
}

func TestDeleteMessage_InvalidID(t *testing.T) {
	h := newTestHandlers(t)

	req := httptest.NewRequest(http.MethodDelete, "/messages/abc", nil)
	ctx := routeContext(req, "id", "abc")
	w := httptest.NewRecorder()

	h.MessageByIDHandler(w, req.WithContext(ctx))

	assertStatusCode(t, w, http.StatusBadRequest)
}

func TestDeleteMessage_NegativeID(t *testing.T) {
	h := newTestHandlers(t)

	req := httptest.NewRequest(http.MethodDelete, "/messages/-1", nil)
	ctx := routeContext(req, "id", "-1")
	w := httptest.NewRecorder()

	h.MessageByIDHandler(w, req.WithContext(ctx))

	assertStatusCode(t, w, http.StatusBadRequest)
}

// routeKey is a custom type for context keys to avoid collisions.
type routeKey string

// routeContext sets path values on the request context (mimics mux routing).
func routeContext(r *http.Request, key, value string) context.Context {
	ctx := r.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, routeKey("route:"+key), value)
}
