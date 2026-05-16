package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	gomvchttp "github.com/mahavirnahata/gophant/http"
)

func TestCORSAllowAll(t *testing.T) {
	r := gomvchttp.NewRouter(nil)
	r.Use(CORS(DefaultCORSConfig()))
	r.Get("/api", func(c *gomvchttp.Context) { c.Text(200, "ok") })

	req := httptest.NewRequest(http.MethodGet, "/api", nil)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Fatalf("expected *, got %q", got)
	}
}

func TestCORSPreflight(t *testing.T) {
	r := gomvchttp.NewRouter(nil)
	r.Use(CORS(DefaultCORSConfig()))
	r.Get("/api", func(c *gomvchttp.Context) { c.Text(200, "ok") })

	req := httptest.NewRequest(http.MethodOptions, "/api", nil)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Access-Control-Request-Method", "GET")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
	if got := w.Header().Get("Access-Control-Allow-Methods"); got == "" {
		t.Fatal("expected Access-Control-Allow-Methods header")
	}
}

func TestCORSSpecificOrigin(t *testing.T) {
	cfg := CORSConfig{
		AllowOrigins: []string{"https://trusted.com"},
		AllowMethods: []string{"GET"},
	}
	r := gomvchttp.NewRouter(nil)
	r.Use(CORS(cfg))
	r.Get("/api", func(c *gomvchttp.Context) { c.Text(200, "ok") })

	// Allowed origin
	req := httptest.NewRequest(http.MethodGet, "/api", nil)
	req.Header.Set("Origin", "https://trusted.com")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "https://trusted.com" {
		t.Fatalf("expected https://trusted.com, got %q", got)
	}

	// Blocked origin
	req2 := httptest.NewRequest(http.MethodGet, "/api", nil)
	req2.Header.Set("Origin", "https://evil.com")
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if got := w2.Header().Get("Access-Control-Allow-Origin"); got == "https://evil.com" {
		t.Fatal("should not have allowed evil.com")
	}
}

func TestCSRFExemptPrefix(t *testing.T) {
	r := gomvchttp.NewRouter(nil)
	r.Use(CSRF(CSRFConfig{
		Secret: []byte("secret"),
		Except: []string{"/api"},
	}))
	r.Post("/api/login", func(c *gomvchttp.Context) { c.Text(200, "ok") })
	r.Post("/web/form", func(c *gomvchttp.Context) { c.Text(200, "ok") })

	// /api/login is exempt — no token needed
	req := httptest.NewRequest(http.MethodPost, "/api/login", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("expected 200 for exempt route, got %d", w.Code)
	}

	// /web/form is NOT exempt — should be blocked
	req2 := httptest.NewRequest(http.MethodPost, "/web/form", nil)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for non-exempt route, got %d", w2.Code)
	}
}

func TestCSRFSkipFunc(t *testing.T) {
	r := gomvchttp.NewRouter(nil)
	r.Use(CSRF(CSRFConfig{
		Secret: []byte("secret"),
		Skip: func(c *gomvchttp.Context) bool {
			return c.GetHeader("X-API-Key") == "supersecret"
		},
	}))
	r.Post("/api/data", func(c *gomvchttp.Context) { c.Text(200, "ok") })

	req := httptest.NewRequest(http.MethodPost, "/api/data", nil)
	req.Header.Set("X-API-Key", "supersecret")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
