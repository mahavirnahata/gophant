package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTo_Redirects(t *testing.T) {
	w := httptest.NewRecorder()
	c := makeCtx(w, httptest.NewRequest(http.MethodGet, "/", nil))
	c.To("/dashboard").Send()

	if w.Code != http.StatusFound {
		t.Fatalf("expected 302, got %d", w.Code)
	}
	if loc := w.Header().Get("Location"); loc != "/dashboard" {
		t.Fatalf("expected /dashboard, got %q", loc)
	}
}

func TestTo_WithStatus(t *testing.T) {
	w := httptest.NewRecorder()
	c := makeCtx(w, httptest.NewRequest(http.MethodGet, "/", nil))
	c.To("/new").WithStatus(http.StatusMovedPermanently).Send()
	if w.Code != http.StatusMovedPermanently {
		t.Fatalf("expected 301, got %d", w.Code)
	}
}

func TestTo_WithFlash(t *testing.T) {
	w := httptest.NewRecorder()
	c := makeCtx(w, httptest.NewRequest(http.MethodGet, "/", nil))
	// Flash is a no-op without a session; just verify no panic and redirect works.
	c.To("/home").With("success", "Saved").Send()
	if w.Code != http.StatusFound {
		t.Fatalf("expected 302, got %d", w.Code)
	}
}

func TestTo_WithError(t *testing.T) {
	w := httptest.NewRecorder()
	c := makeCtx(w, httptest.NewRequest(http.MethodGet, "/", nil))
	c.To("/login").WithError("Unauthorized").Send()
	if w.Code != http.StatusFound {
		t.Fatalf("expected 302, got %d", w.Code)
	}
}

func TestTo_WithSuccess(t *testing.T) {
	w := httptest.NewRecorder()
	c := makeCtx(w, httptest.NewRequest(http.MethodGet, "/", nil))
	c.To("/home").WithSuccess("Done!").Send()
	if w.Code != http.StatusFound {
		t.Fatalf("expected 302, got %d", w.Code)
	}
}

func TestTo_WithErrors(t *testing.T) {
	w := httptest.NewRecorder()
	c := makeCtx(w, httptest.NewRequest(http.MethodGet, "/", nil))
	c.To("/form").WithErrors(map[string][]string{"email": {"required"}}).Send()
	if w.Code != http.StatusFound {
		t.Fatalf("expected 302, got %d", w.Code)
	}
}

func TestTo_WithInput(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Form = map[string][]string{"name": {"Alice"}}
	w := httptest.NewRecorder()
	c := makeCtx(w, req)
	c.To("/form").WithInput().Send()
	if w.Code != http.StatusFound {
		t.Fatalf("expected 302, got %d", w.Code)
	}
}

func TestOldInput_Empty(t *testing.T) {
	w := httptest.NewRecorder()
	c := makeCtx(w, httptest.NewRequest(http.MethodGet, "/", nil))
	if c.OldInput("name") != "" {
		t.Fatal("expected empty OldInput when no flash")
	}
}
