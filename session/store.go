package session

import "sync"

type Store interface {
	Get(id string) (map[string]any, bool)
	Save(id string, values map[string]any) error
	Delete(id string) error
}

type MemoryStore struct {
	mu    sync.RWMutex
	items map[string]map[string]any
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{items: map[string]map[string]any{}}
}

func (s *MemoryStore) Get(id string) (map[string]any, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	vals, ok := s.items[id]
	if !ok {
		return nil, false
	}
	copy := map[string]any{}
	for k, v := range vals {
		copy[k] = v
	}
	return copy, true
}

func (s *MemoryStore) Save(id string, values map[string]any) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	copy := map[string]any{}
	for k, v := range values {
		copy[k] = v
	}
	s.items[id] = copy
	return nil
}

func (s *MemoryStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.items, id)
	return nil
}
