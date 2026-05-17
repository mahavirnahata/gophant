package http

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSSE_SetsHeaders(t *testing.T) {
	w := httptest.NewRecorder()
	c := makeCtx(w, httptest.NewRequest(http.MethodGet, "/events", nil))
	sse := c.SSE()
	if sse == nil {
		t.Fatal("SSE() should return non-nil for httptest.ResponseRecorder")
	}
	if ct := w.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/event-stream") {
		t.Fatalf("expected text/event-stream, got %q", ct)
	}
	if w.Header().Get("Cache-Control") != "no-cache" {
		t.Fatal("expected Cache-Control: no-cache")
	}
}

func TestSSE_SendString(t *testing.T) {
	w := httptest.NewRecorder()
	c := makeCtx(w, httptest.NewRequest(http.MethodGet, "/", nil))
	sse := c.SSE()
	if err := sse.SendString("hello world"); err != nil {
		t.Fatalf("SendString: %v", err)
	}
	body := w.Body.String()
	if !strings.Contains(body, "data: hello world") {
		t.Fatalf("expected 'data: hello world' in body, got: %q", body)
	}
}

func TestSSE_SendWithEvent(t *testing.T) {
	w := httptest.NewRecorder()
	c := makeCtx(w, httptest.NewRequest(http.MethodGet, "/", nil))
	sse := c.SSE()
	sse.Send(SSEEvent{Event: "update", Data: "payload"})
	body := w.Body.String()
	if !strings.Contains(body, "event: update") {
		t.Fatalf("expected 'event: update', got: %q", body)
	}
	if !strings.Contains(body, "data: payload") {
		t.Fatalf("expected 'data: payload', got: %q", body)
	}
}

func TestSSE_SendWithID(t *testing.T) {
	w := httptest.NewRecorder()
	c := makeCtx(w, httptest.NewRequest(http.MethodGet, "/", nil))
	sse := c.SSE()
	sse.Send(SSEEvent{ID: "42", Data: "msg"})
	body := w.Body.String()
	if !strings.Contains(body, "id: 42") {
		t.Fatalf("expected 'id: 42', got: %q", body)
	}
}

func TestSSE_SendJSON(t *testing.T) {
	w := httptest.NewRecorder()
	c := makeCtx(w, httptest.NewRequest(http.MethodGet, "/", nil))
	sse := c.SSE()
	sse.SendJSON(map[string]any{"type": "ping"})
	body := w.Body.String()
	if !strings.Contains(body, `"type":"ping"`) {
		t.Fatalf("expected JSON data, got: %q", body)
	}
}

func TestSSE_Comment(t *testing.T) {
	w := httptest.NewRecorder()
	c := makeCtx(w, httptest.NewRequest(http.MethodGet, "/", nil))
	sse := c.SSE()
	sse.Comment("keep-alive")
	body := w.Body.String()
	if !strings.Contains(body, ": keep-alive") {
		t.Fatalf("expected comment in body, got: %q", body)
	}
}

func TestSSE_Retry(t *testing.T) {
	w := httptest.NewRecorder()
	c := makeCtx(w, httptest.NewRequest(http.MethodGet, "/", nil))
	sse := c.SSE()
	sse.Retry(3000)
	body := w.Body.String()
	if !strings.Contains(body, "retry: 3000") {
		t.Fatalf("expected 'retry: 3000', got: %q", body)
	}
}
