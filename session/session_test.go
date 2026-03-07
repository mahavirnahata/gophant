package session

import (
	"net/http/httptest"
	"testing"
)

func TestSessionSetGet(t *testing.T) {
	m := NewManager([]byte("secret"))
	req := httptest.NewRequest("GET", "/", nil)
	sess, _ := m.load(req)
	sess.Set("foo", "bar")
	if v, ok := sess.Get("foo"); !ok || v != "bar" {
		t.Fatalf("expected value")
	}
}

func TestSessionRegenerate(t *testing.T) {
	m := NewManager([]byte("secret"))
	req := httptest.NewRequest("GET", "/", nil)
	sess, _ := m.load(req)
	old := sess.ID
	sess.Regenerate()
	if sess.ID == old {
		t.Fatalf("expected new session id")
	}
}
