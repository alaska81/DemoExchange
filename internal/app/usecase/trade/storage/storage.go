package storage

import (
	"fmt"
	"sync"

	"DemoExchange/internal/app/entities"
)

type Storage struct {
	orders map[string]*entities.Order
	mu     sync.RWMutex
}

func New() *Storage {
	return &Storage{
		orders: make(map[string]*entities.Order),
		mu:     sync.RWMutex{},
	}
}

func (s *Storage) Set(order *entities.Order) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.orders[order.OrderUID] = order
	fmt.Println("Add: ", order.OrderUID)
}

func (s *Storage) Get(orderUID string) (*entities.Order, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	order, ok := s.orders[orderUID]

	fmt.Println("Get: ", orderUID)
	return order, ok
}

func (s *Storage) Delete(orderUID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.orders, orderUID)
	fmt.Println("Delete: ", orderUID)
}

func (s *Storage) List() []*entities.Order {
	s.mu.RLock()
	defer s.mu.RUnlock()

	orders := make([]*entities.Order, 0, len(s.orders))
	for _, order := range s.orders {
		orders = append(orders, order)
	}

	return orders
}
