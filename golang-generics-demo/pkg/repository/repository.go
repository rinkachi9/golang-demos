package repository

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

var ErrNotFound = errors.New("entity not found")

// Entity is the constraint for objects stored in the repository.
// They must have a comparable ID.
type Entity[ID comparable] interface {
	GetID() ID
}

// Repository defines standard CRUD operations.
type Repository[T Entity[ID], ID comparable] interface {
	Create(ctx context.Context, entity T) error
	FindByID(ctx context.Context, id ID) (T, error)
	FindAll(ctx context.Context) ([]T, error)
	Update(ctx context.Context, entity T) error
	Delete(ctx context.Context, id ID) error
}

// InMemoryRepository is a thread-safe implementation.
type InMemoryRepository[T Entity[ID], ID comparable] struct {
	mu   sync.RWMutex
	data map[ID]T
}

func NewInMemoryRepository[T Entity[ID], ID comparable]() *InMemoryRepository[T, ID] {
	return &InMemoryRepository[T, ID]{
		data: make(map[ID]T),
	}
}

func (r *InMemoryRepository[T, ID]) Create(ctx context.Context, entity T) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	id := entity.GetID()
	if _, exists := r.data[id]; exists {
		return fmt.Errorf("entity with ID %v already exists", id)
	}
	r.data[id] = entity
	return nil
}

func (r *InMemoryRepository[T, ID]) FindByID(ctx context.Context, id ID) (T, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entity, ok := r.data[id]
	if !ok {
		var zero T
		return zero, ErrNotFound
	}
	return entity, nil
}

func (r *InMemoryRepository[T, ID]) FindAll(ctx context.Context) ([]T, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]T, 0, len(r.data))
	for _, item := range r.data {
		result = append(result, item)
	}
	return result, nil
}

func (r *InMemoryRepository[T, ID]) Update(ctx context.Context, entity T) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	id := entity.GetID()
	if _, exists := r.data[id]; !exists {
		return ErrNotFound
	}
	r.data[id] = entity
	return nil
}

func (r *InMemoryRepository[T, ID]) Delete(ctx context.Context, id ID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.data[id]; !exists {
		return ErrNotFound
	}
	delete(r.data, id)
	return nil
}
