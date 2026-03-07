package schema

import "testing"

func TestSchemaBuild(t *testing.T) {
	b := New("mysql")
	bp, sql := b.Build("users", func(t *Blueprint) {
		t.Increments("id")
		t.String("email", 255)
		t.Unique("email")
		t.Index("email")
	})
	if sql == "" {
		t.Fatalf("expected sql")
	}
	idx := b.Indexes(bp)
	if len(idx) != 1 {
		t.Fatalf("expected index")
	}
}
