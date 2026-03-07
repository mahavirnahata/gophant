package cache

import (
	"testing"
	"time"
)

func TestMemoryCache(t *testing.T) {
	m := NewMemoryStore()
	c := New(m)

	_ = c.Set("a", "b", time.Minute)
	if v, ok := c.Get("a"); !ok || v.(string) != "b" {
		t.Fatalf("expected cached value")
	}

	_ = c.FlushTag("missing")
	key := c.Tag("users", "list")
	_ = c.Set(key, "x", time.Minute)
	_ = c.FlushTag("users")
	if _, ok := c.Get(key); ok {
		t.Fatalf("expected tag flush")
	}
}
