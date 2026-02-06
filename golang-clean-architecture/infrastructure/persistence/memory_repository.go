package persistence

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rinkachi/golang-demos/golang-clean-architecture/domain"
)

type InMemoryOrderRepository struct {
	mu     sync.RWMutex
	orders map[uuid.UUID]*domain.Order
}

func NewInMemoryOrderRepository() *InMemoryOrderRepository {
	return &InMemoryOrderRepository{
		orders: make(map[uuid.UUID]*domain.Order),
	}
}

func (r *InMemoryOrderRepository) Save(ctx context.Context, order *domain.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Simulate IO delay
	time.Sleep(10 * time.Millisecond)

	r.orders[order.ID] = order
	return nil
}

func (r *InMemoryOrderRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Simulate IO delay
	time.Sleep(5 * time.Millisecond)

	order, ok := r.orders[id]
	if !ok {
		return nil, domain.ErrOrderNotFound
	}
	return order, nil
}

func (r *InMemoryOrderRepository) FindAll(ctx context.Context) ([]*domain.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	orders := make([]*domain.Order, 0, len(r.orders))
	for _, order := range r.orders {
		orders = append(orders, order)
	}
	return orders, nil
}
