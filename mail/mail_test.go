package mail

import "testing"

func TestNullDriverCaptures(t *testing.T) {
	drv := NewNullDriver()
	m := New(drv)
	m.From = "no-reply@app.com"

	if err := m.To("alice@example.com").Subject("Hi").Text("Hello Alice").Send(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	drv.AssertSent(t, "alice@example.com")
	if len(drv.Sent) != 1 {
		t.Fatalf("expected 1 message, got %d", len(drv.Sent))
	}
	if drv.Sent[0].Subject != "Hi" {
		t.Fatalf("unexpected subject: %q", drv.Sent[0].Subject)
	}
}

func TestNullDriverReset(t *testing.T) {
	drv := NewNullDriver()
	m := New(drv)
	_ = m.To("x@x.com").Subject("s").Send()
	drv.Reset()
	drv.AssertNotSent(t)
}

func TestChaining(t *testing.T) {
	drv := NewNullDriver()
	m := New(drv)
	err := m.To("bob@example.com", "carol@example.com").
		CC("dave@example.com").
		Subject("Team update").
		HTML("<b>Hi</b>").
		Text("Hi").
		Send()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(drv.Sent[0].To) != 2 {
		t.Fatalf("expected 2 recipients, got %d", len(drv.Sent[0].To))
	}
	if len(drv.Sent[0].CC) != 1 {
		t.Fatalf("expected 1 CC, got %d", len(drv.Sent[0].CC))
	}
}

func TestBuildRFC2822TextOnly(t *testing.T) {
	msg := &Message{From: "a@b.com", To: []string{"c@d.com"}, Subject: "Hello", Text: "body text"}
	out := buildRFC2822("a@b.com", msg)
	if !contains(out, "text/plain") {
		t.Fatalf("expected text/plain content type, got:\n%s", out)
	}
	if !contains(out, "body text") {
		t.Fatalf("expected body in output, got:\n%s", out)
	}
}

func TestBuildRFC2822Multipart(t *testing.T) {
	msg := &Message{From: "a@b.com", To: []string{"c@d.com"}, Subject: "Hi", Text: "text", HTML: "<b>html</b>"}
	out := buildRFC2822("a@b.com", msg)
	if !contains(out, "multipart/alternative") {
		t.Fatalf("expected multipart, got:\n%s", out)
	}
}

func contains(s, sub string) bool {
	return len(sub) == 0 || (len(s) >= len(sub) && func() bool {
		for i := 0; i <= len(s)-len(sub); i++ {
			if s[i:i+len(sub)] == sub {
				return true
			}
		}
		return false
	}())
}
