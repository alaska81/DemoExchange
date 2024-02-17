package cache

import (
	"fmt"
	"sync"
)

type Cache struct {
	values map[string]any
	mu     sync.RWMutex
}

func New() *Cache {
	return &Cache{
		values: make(map[string]any),
		mu:     sync.RWMutex{},
	}
}

func (s *Cache) Set(uid string, value any) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.values[uid] = value
	fmt.Println("Add: ", uid)
}

func (s *Cache) Get(uid string) (any, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	value, ok := s.values[uid]

	fmt.Println("Get: ", uid)
	return value, ok
}

func (s *Cache) Delete(uid string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.values, uid)
	fmt.Println("Delete: ", uid)
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
