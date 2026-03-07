package auth

import (
	"net/http/httptest"
	"testing"

	gomvchttp "github.com/mahavirnahata/gophant/http"
	"github.com/mahavirnahata/gophant/session"
)

func TestAuthLogin(t *testing.T) {
	m := session.NewManager([]byte("secret"))
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	ctx := gomvchttp.NewContext(w, req, nil)
	var sess *session.Session
	m.Middleware()(func(c *gomvchttp.Context) {
		if v, ok := c.Get("session"); ok {
			sess = v.(*session.Session)
		}
	})(ctx)
	ctx.Set("session", sess)

	a := NewManager()
	a.Login(ctx, "1")
	if id, ok := a.UserID(ctx); !ok || id != "1" {
		t.Fatalf("expected user id")
	}
}

func TestRoles(t *testing.T) {
	ctx := gomvchttp.NewContext(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil), nil)
	ctx.Set("session", &session.Session{Values: map[string]any{}})
	m := NewManager()
	m.SetRoles(ctx, []string{"admin"})
	if !m.HasRole(ctx, "admin") {
		t.Fatalf("expected admin role")
	}
}
