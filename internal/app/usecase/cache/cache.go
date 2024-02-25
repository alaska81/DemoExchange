package cache

import (
	"fmt"
	"sync"
)

type Logger interface {
	Info(args ...interface{})
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
func New[K comparable, V any](log Logger) *Cache[K, V] {
	return &Cache[K, V]{
		name:   fmt.Sprintf("%T", *new(V)),
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

// List returns all values from cache
func (c *Cache[K, V]) List() []V {
	c.mu.RLock()
	defer c.mu.RUnlock()

	values := make([]V, 0, len(c.values))

	for _, value := range c.values {
		values = append(values, value)
	}

	return values
}
