package cache

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	gomvchttp "github.com/mahavirnahata/gophant/http"
)

type stubView struct{}

func (s stubView) Render(w io.Writer, name string, data map[string]any) error { return nil }

func TestResponseCache(t *testing.T) {
	c := New(NewMemoryStore())
	r := gomvchttp.NewRouter(stubView{})
	r.Use(ResponseCache(c, ResponseCacheOptions{TTL: time.Minute}))
	r.Get("/", func(ctx *gomvchttp.Context) {
		ctx.Text(200, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req)

	if w2.Body.String() != "ok" {
		t.Fatalf("expected cached response")
	}
}
