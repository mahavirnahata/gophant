package http

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// ── Group middleware ─────────────────────────────────────────────────────────

func TestGroupMiddleware(t *testing.T) {
	r := NewRouter(stubViewRouter{})

	called := false
	mw := func(next Handler) Handler {
		return func(c *Context) {
			called = true
			next(c)
		}
	}

	r.Group("/api", func(rg *Router) {
		rg.Get("/ping", func(c *Context) { c.Text(200, "pong") })
	}, mw)

	req := httptest.NewRequest(http.MethodGet, "/api/ping", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if !called {
		t.Fatal("group middleware was not called")
	}
}

func TestGroupMiddlewareNotAppliedOutsideGroup(t *testing.T) {
	r := NewRouter(stubViewRouter{})

	called := false
	mw := func(next Handler) Handler {
		return func(c *Context) {
			called = true
			next(c)
		}
	}

	r.Get("/outside", func(c *Context) { c.Text(200, "outside") })
	r.Group("/api", func(rg *Router) {
		rg.Get("/ping", func(c *Context) { c.Text(200, "ping") })
	}, mw)

	req := httptest.NewRequest(http.MethodGet, "/outside", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if called {
		t.Fatal("group middleware should not apply to routes outside the group")
	}
}

// ── Form method spoofing ─────────────────────────────────────────────────────

func TestFormMethodSpoofPUT(t *testing.T) {
	r := NewRouter(stubViewRouter{})
	r.Put("/users/{id}", func(c *Context) {
		c.Text(200, "updated:"+c.Param("id"))
	})

	body := strings.NewReader("_method=PUT")
	req := httptest.NewRequest(http.MethodPost, "/users/42", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if w.Body.String() != "updated:42" {
		t.Fatalf("unexpected body: %s", w.Body.String())
	}
}

func TestFormMethodSpoofDELETE(t *testing.T) {
	r := NewRouter(stubViewRouter{})
	r.Delete("/items/{id}", func(c *Context) {
		c.Text(200, "deleted:"+c.Param("id"))
	})

	body := strings.NewReader("_method=DELETE")
	req := httptest.NewRequest(http.MethodPost, "/items/7", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

// ── Named routes and Routes() ────────────────────────────────────────────────

func TestNamedRouteURL(t *testing.T) {
	r := NewRouter(stubViewRouter{})
	r.Get("/users/{id}", func(c *Context) {}).Name("users.show")

	got := r.URL("users.show", "99")
	if got != "/users/99" {
		t.Fatalf("expected /users/99, got %q", got)
	}
}

func TestRoutesMeta(t *testing.T) {
	r := NewRouter(stubViewRouter{})
	r.Get("/ping", func(c *Context) {}).Name("ping")
	r.Post("/users", func(c *Context) {}).Name("users.store")

	routes := r.Routes()
	if len(routes) != 2 {
		t.Fatalf("expected 2 routes, got %d", len(routes))
	}
	found := false
	for _, rt := range routes {
		if rt.Name == "users.store" && rt.Method == http.MethodPost && rt.Pattern == "/users" {
			found = true
		}
	}
	if !found {
		t.Fatal("users.store route not found in Routes()")
	}
}

// ── Abort ────────────────────────────────────────────────────────────────────

func TestAbort(t *testing.T) {
	r := NewRouter(stubViewRouter{})
	r.Use(func(next Handler) Handler {
		return func(c *Context) {
			c.Abort()
			c.Text(401, "unauthorized")
		}
	})
	r.Get("/secret", func(c *Context) {
		c.Text(200, "secret")
	})

	req := httptest.NewRequest(http.MethodGet, "/secret", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 401 {
		t.Fatalf("expected 401, got %d", w.Code)
	}
	if w.Body.String() != "unauthorized" {
		t.Fatalf("expected unauthorized, got %q", w.Body.String())
	}
}

// ── Cookie helpers ───────────────────────────────────────────────────────────

func TestSetAndReadCookie(t *testing.T) {
	r := NewRouter(stubViewRouter{})
	r.Get("/set", func(c *Context) {
		c.SetCookie(&http.Cookie{Name: "x", Value: "42"})
		c.Text(200, "ok")
	})
	r.Get("/read", func(c *Context) {
		ck, err := c.Cookie("x")
		if err != nil {
			c.Text(400, "missing")
			return
		}
		c.Text(200, ck.Value)
	})

	// Set cookie
	req := httptest.NewRequest(http.MethodGet, "/set", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Header().Get("Set-Cookie") == "" {
		t.Fatal("expected Set-Cookie header")
	}

	// Read it back
	req2 := httptest.NewRequest(http.MethodGet, "/read", nil)
	req2.AddCookie(&http.Cookie{Name: "x", Value: "42"})
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if w2.Body.String() != "42" {
		t.Fatalf("expected cookie value 42, got %q", w2.Body.String())
	}
}

// ── Back() ───────────────────────────────────────────────────────────────────

func TestBack(t *testing.T) {
	r := NewRouter(stubViewRouter{})
	r.Post("/go-back", func(c *Context) {
		c.Back("/default")
	})

	req := httptest.NewRequest(http.MethodPost, "/go-back", nil)
	req.Header.Set("Referer", "/previous-page")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusFound {
		t.Fatalf("expected 302, got %d", w.Code)
	}
	if loc := w.Header().Get("Location"); loc != "/previous-page" {
		t.Fatalf("expected /previous-page, got %q", loc)
	}
}

func TestBackFallback(t *testing.T) {
	r := NewRouter(stubViewRouter{})
	r.Post("/go-back", func(c *Context) {
		c.Back("/fallback")
	})

	req := httptest.NewRequest(http.MethodPost, "/go-back", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if loc := w.Header().Get("Location"); loc != "/fallback" {
		t.Fatalf("expected /fallback, got %q", loc)
	}
}
