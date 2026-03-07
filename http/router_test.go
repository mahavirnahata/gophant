package http

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

type stubViewRouter struct{}

func (s stubViewRouter) Render(w io.Writer, name string, data map[string]any) error {
	return nil
}

func TestRouterParamsAndMethods(t *testing.T) {
	r := NewRouter(stubViewRouter{})
	r.Get("/users/{id}", func(c *Context) {
		c.Text(200, c.Param("id"))
	})

	req := httptest.NewRequest(http.MethodGet, "/users/42", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if w.Body.String() != "42" {
		t.Fatalf("expected body 42, got %q", w.Body.String())
	}
}

func TestRouterMethodNotAllowed(t *testing.T) {
	r := NewRouter(stubViewRouter{})
	r.Get("/ping", func(c *Context) { c.Text(200, "ok") })

	req := httptest.NewRequest(http.MethodPost, "/ping", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestRouterGroup(t *testing.T) {
	r := NewRouter(stubViewRouter{})
	r.Group("/admin", func(rg *Router) {
		rg.Get("/dashboard", func(c *Context) { c.Text(200, "dash") })
	})

	req := httptest.NewRequest(http.MethodGet, "/admin/dashboard", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
