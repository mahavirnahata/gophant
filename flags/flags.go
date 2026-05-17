// Package flags provides a simple feature-flag system.
//
// Flags can be toggled at runtime without a deploy. The in-memory store is
// the default; swap it for a DB-backed or Redis-backed store in production.
//
//	flags.Enable("new-checkout")
//	if flags.IsEnabled("new-checkout") { ... }
package flags

import "sync"

// Store is the interface for a feature-flag backend.
type Store interface {
	IsEnabled(feature string) bool
	Enable(feature string)
	Disable(feature string)
	All() map[string]bool
}

// MemoryStore is the default in-memory store. It is safe for concurrent use.
type MemoryStore struct {
	mu   sync.RWMutex
	data map[string]bool
}

// NewMemoryStore returns an empty in-memory feature-flag store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{data: map[string]bool{}}
}

func (s *MemoryStore) IsEnabled(feature string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data[feature]
}

func (s *MemoryStore) Enable(feature string) {
	s.mu.Lock()
	s.data[feature] = true
	s.mu.Unlock()
}

func (s *MemoryStore) Disable(feature string) {
	s.mu.Lock()
	s.data[feature] = false
	s.mu.Unlock()
}

func (s *MemoryStore) All() map[string]bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make(map[string]bool, len(s.data))
	for k, v := range s.data {
		out[k] = v
	}
	return out
}

// ── package-level default store ───────────────────────────────────────────────

var defaultStore Store = NewMemoryStore()

// SetStore replaces the package-level store (e.g., to use a DB-backed store).
func SetStore(s Store) { defaultStore = s }

// IsEnabled reports whether feature is enabled in the default store.
func IsEnabled(feature string) bool { return defaultStore.IsEnabled(feature) }

// Enable enables feature in the default store.
func Enable(feature string) { defaultStore.Enable(feature) }

// Disable disables feature in the default store.
func Disable(feature string) { defaultStore.Disable(feature) }

// All returns all flags and their states from the default store.
func All() map[string]bool { return defaultStore.All() }

// When calls fn only when feature is enabled. Useful for inline branching:
//
//	flags.When("dark-mode", func() { enableDarkMode() })
func When(feature string, fn func()) {
	if IsEnabled(feature) {
		fn()
	}
}

// Toggle flips the state of a feature flag in the default store.
func Toggle(feature string) {
	if IsEnabled(feature) {
		Disable(feature)
	} else {
		Enable(feature)
	}
}
