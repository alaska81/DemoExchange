package cache

import (
	"fmt"
	"sync"
)

// type Values[K comparable, V any] map[K]V

type Cache struct {
	name   string
	values map[string]any
	mu     sync.RWMutex
}

func New(name string) *Cache {
	return &Cache{
		name:   name,
		values: make(map[string]any),
		mu:     sync.RWMutex{},
	}
}

func (s *Cache) Set(uid string, value any) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.values[uid] = value
	fmt.Printf("Cache [%s] Add: %s\n", s.name, uid)
}

func (s *Cache) Get(uid string) (any, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	value, ok := s.values[uid]

	fmt.Printf("Cache [%s] Get: %s\n", s.name, uid)
	return value, ok
}

func (s *Cache) Delete(uid string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.values, uid)
	fmt.Printf("Cache [%s] Delete: %s\n", s.name, uid)
}

func (s *Cache) List() []any {
	s.mu.RLock()
	defer s.mu.RUnlock()

	values := make([]any, 0, len(s.values))
	for _, vakue := range s.values {
		values = append(values, vakue)
	}

	return values
}
