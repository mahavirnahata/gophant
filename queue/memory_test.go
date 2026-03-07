package queue

import "testing"

func TestMemoryQueue(t *testing.T) {
	q := NewMemoryQueue(1)

	called := false
	_ = q.Push(func() error {
		called = true
		return nil
	})

	go func() { _ = q.Run() }()
	_ = q.Close()

	if !called {
		t.Fatalf("expected job to run")
	}
}
