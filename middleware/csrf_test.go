package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	gomvchttp "github.com/mahavirnahata/gophant/http"
)

func TestCSRFMiddlewareBlocksMissing(t *testing.T) {
	r := gomvchttp.NewRouter(nil)
	r.Use(CSRF(CSRFConfig{Secret: []byte("secret")}))
	r.Post("/submit", func(c *gomvchttp.Context) {
		c.Text(200, "ok")
	})

	req := httptest.NewRequest(http.MethodPost, "/submit", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
}
