package events

import (
	"reflect"
	"sync"
	"testing"
)

type UserRegistered struct{ Email string }
type OrderPlaced struct{ ID int }

func TestListenAndFire(t *testing.T) {
	Flush()
	var got string
	Listen(func(e UserRegistered) { got = e.Email })
	Fire(UserRegistered{Email: "alice@example.com"})
	if got != "alice@example.com" {
		t.Fatalf("expected alice@example.com, got %q", got)
	}
}

func TestMultipleListeners(t *testing.T) {
	Flush()
	count := 0
	Listen(func(e UserRegistered) { count++ })
	Listen(func(e UserRegistered) { count++ })
	Fire(UserRegistered{})
	if count != 2 {
		t.Fatalf("expected 2 listeners called, got %d", count)
	}
}

func TestFireDoesNotCrossTypes(t *testing.T) {
	Flush()
	called := false
	Listen(func(e OrderPlaced) { called = true })
	Fire(UserRegistered{})
	if called {
		t.Fatal("OrderPlaced listener should not fire for UserRegistered event")
	}
}

func TestFireAsync(t *testing.T) {
	Flush()
	var wg sync.WaitGroup
	wg.Add(1)
	Listen(func(e UserRegistered) { wg.Done() })
	FireAsync(UserRegistered{Email: "bob@example.com"})
	wg.Wait()
}

func TestIsolatedBus(t *testing.T) {
	b := NewBus()
	called := false
	b.On(reflect.TypeOf(UserRegistered{}), func(e any) { called = true })
	b.Fire(UserRegistered{})
	if !called {
		t.Fatal("isolated bus listener not called")
	}
}

func TestFlushClearsListeners(t *testing.T) {
	Flush()
	called := false
	Listen(func(e UserRegistered) { called = true })
	Flush()
	Fire(UserRegistered{})
	if called {
		t.Fatal("listener should have been removed after Flush()")
	}
}
