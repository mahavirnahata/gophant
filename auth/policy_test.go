package auth

import (
	"net/http/httptest"
	"testing"

	gomvchttp "github.com/mahavirnahata/gophant/http"
)

type testPolicy struct{}

func (p *testPolicy) View(c *gomvchttp.Context) bool { return true }

func TestPolicyRegistrar(t *testing.T) {
	g := NewGate()
	reg := NewPolicyRegistrar(g)
	reg.Register("item", &testPolicy{})

	ctx := gomvchttp.NewContext(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil), nil)
	if !g.Allows("item.view", ctx) {
		t.Fatalf("expected policy to allow")
	}
}
