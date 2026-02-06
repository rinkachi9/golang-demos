package domain

import (
	"context"

	"github.com/google/uuid"
)

// OrderRepository defines the interface for persisting orders.
// It follows the dependency inversion principle.
type OrderRepository interface {
	Save(ctx context.Context, order *Order) error
	FindByID(ctx context.Context, id uuid.UUID) (*Order, error)
	FindAll(ctx context.Context) ([]*Order, error)
}
