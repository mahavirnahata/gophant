// Package events provides a zero-dependency in-process event bus with generic
// typed listeners.
//
// Usage:
//
//	// Register a listener (usually in init or main):
//	events.Listen(func(e UserRegistered) {
//	    mailer.To(e.Email).Subject("Welcome!").Send()
//	})
//
//	// Fire an event from a controller or service:
//	events.Fire(UserRegistered{Email: "alice@example.com"})
//
//	// Fire without blocking the current goroutine:
//	events.FireAsync(OrderPlaced{OrderID: 42})
package events

import (
	"reflect"
	"sync"
)

// Bus is a typed event bus. The package-level functions use a shared default bus.
type Bus struct {
	mu        sync.RWMutex
	listeners map[reflect.Type][]func(any)
}

// NewBus creates an isolated event bus.
func NewBus() *Bus {
	return &Bus{listeners: map[reflect.Type][]func(any){}}
}

// On registers a raw listener for the given reflect.Type on this bus.
func (b *Bus) On(t reflect.Type, fn func(any)) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.listeners[t] = append(b.listeners[t], fn)
}

// Fire dispatches event to all matching listeners synchronously.
func (b *Bus) Fire(event any) {
	b.mu.RLock()
	fns := append([]func(any){}, b.listeners[reflect.TypeOf(event)]...)
	b.mu.RUnlock()
	for _, fn := range fns {
		fn(event)
	}
}

// FireAsync dispatches event in a new goroutine.
func (b *Bus) FireAsync(event any) {
	go b.Fire(event)
}

// Flush removes all listeners (useful in tests).
func (b *Bus) Flush() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.listeners = map[reflect.Type][]func(any){}
}

// ── Default bus ───────────────────────────────────────────────────────────────

var defaultBus = NewBus()

// Listen registers a typed listener on the default bus.
//
//	events.Listen(func(e UserRegistered) { ... })
func Listen[E any](fn func(E)) {
	t := reflect.TypeOf((*E)(nil)).Elem()
	defaultBus.On(t, func(e any) { fn(e.(E)) })
}

// Fire dispatches event to all matching listeners on the default bus.
func Fire(event any) { defaultBus.Fire(event) }

// FireAsync dispatches event asynchronously on the default bus.
func FireAsync(event any) { defaultBus.FireAsync(event) }

// Flush clears all listeners on the default bus (useful in tests).
func Flush() { defaultBus.Flush() }
