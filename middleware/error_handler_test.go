package middleware

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	gomvchttp "github.com/mahavirnahata/gophant/http"
)

type stubView struct{}

func (s stubView) Render(w io.Writer, name string, data map[string]any) error { return nil }

func TestErrorHandler(t *testing.T) {
	r := gomvchttp.NewRouter(stubView{})
	r.Use(ErrorHandler(nil))
	r.Get("/", func(c *gomvchttp.Context) {
		c.Error(errors.New("boom"))
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}
