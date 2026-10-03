package cache

import (
	"fmt"
	"sync"
)

type Logger interface {
	Info(args ...any)
}

type Values[K comparable, V any] map[K]V

// Cache is a generic in-memory cache
type Cache[K comparable, V any] struct {
	name   string
	mu     sync.RWMutex
	values Values[K, V]
	log    Logger
}

// New creates a new instance of Cache
func New[K comparable, V any](name string, log Logger) *Cache[K, V] {
	return &Cache[K, V]{
		name:   name,
		mu:     sync.RWMutex{},
		values: make(Values[K, V]),
		log:    log,
	}
}

// Set adds value to cache
func (c *Cache[K, V]) Set(uid K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.values[uid] = value
	c.log.Info(fmt.Sprintf("Cache [%s] Set: %v", c.name, uid))
}

// Get retrieves a value from cache by uid
func (c *Cache[K, V]) Get(uid K) (value V, ok bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	value, ok = c.values[uid]
	return
}

// Delete removes a value from cache by uid
func (c *Cache[K, V]) Delete(uid K) {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, ok := c.values[uid]
	if !ok {
		return
	}

	delete(c.values, uid)
	c.log.Info(fmt.Sprintf("Cache [%s] Delete: %v", c.name, uid))
}
