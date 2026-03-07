package cache

import (
	"sync"
	"time"
)

type item struct {
	value any
	exp   time.Time
}

type MemoryStore struct {
	mu    sync.RWMutex
	items map[string]item
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{items: map[string]item{}}
}

func (m *MemoryStore) Get(key string) (any, bool) {
	m.mu.RLock()
	it, ok := m.items[key]
	m.mu.RUnlock()
	if !ok {
		return nil, false
	}
	if !it.exp.IsZero() && time.Now().After(it.exp) {
		_ = m.Delete(key)
		return nil, false
	}
	return it.value, true
}

func (m *MemoryStore) Set(key string, value any, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	it := item{value: value}
	if ttl > 0 {
		it.exp = time.Now().Add(ttl)
	}
	m.items[key] = it
	return nil
}

func (m *MemoryStore) Delete(key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.items, key)
	return nil
}

func (m *MemoryStore) Flush() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.items = map[string]item{}
	return nil
}

func (m *MemoryStore) FlushTag(tag string) error {
	if tag == "" {
		return nil
	}
	prefix := tagKey(tag, "")
	m.mu.Lock()
	defer m.mu.Unlock()
	for k := range m.items {
		if len(k) >= len(prefix) && k[:len(prefix)] == prefix {
			delete(m.items, k)
		}
	}
	return nil
}
