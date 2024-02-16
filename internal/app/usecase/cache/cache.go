package cache

import (
	"fmt"
	"sync"

	"DemoExchange/internal/app/entities"
)

type Cache struct {
	orders map[string]*entities.Order
	mu     sync.RWMutex
}

func New() *Cache {
	return &Cache{
		orders: make(map[string]*entities.Order),
		mu:     sync.RWMutex{},
	}
}

func (s *Cache) Set(order *entities.Order) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.orders[order.OrderUID] = order
	fmt.Println("Add: ", order.OrderUID)
}

func (s *Cache) Get(orderUID string) (*entities.Order, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	order, ok := s.orders[orderUID]

	fmt.Println("Get: ", orderUID)
	return order, ok
}

func (s *Cache) Delete(orderUID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.orders, orderUID)
	fmt.Println("Delete: ", orderUID)
}

func (s *Cache) List() []*entities.Order {
	s.mu.RLock()
	defer s.mu.RUnlock()

	orders := make([]*entities.Order, 0, len(s.orders))
	for _, order := range s.orders {
		orders = append(orders, order)
	}

	return orders
}
