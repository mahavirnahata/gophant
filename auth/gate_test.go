package auth

import (
	"net/http/httptest"
	"testing"

	gomvchttp "github.com/mahavirnahata/gophant/http"
)

func TestGate(t *testing.T) {
	g := NewGate()
	g.Define("view", func(c *gomvchttp.Context) bool { return true })
	ctx := gomvchttp.NewContext(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil), nil)
	if !g.Allows("view", ctx) {
		t.Fatalf("expected allowed")
	}
}
