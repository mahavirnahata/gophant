package security

import (
	"net/http/httptest"
	"testing"
)

func TestApplyHeaders(t *testing.T) {
	w := httptest.NewRecorder()
	ApplyHeaders(w, DefaultHeaders())
	if w.Header().Get("X-Content-Type-Options") == "" {
		t.Fatalf("expected header")
	}
}
