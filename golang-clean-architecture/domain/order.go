package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Sentinel errors for the domain
var (
	ErrOrderNotFound    = errors.New("order not found")
	ErrProductNotFound  = errors.New("product not found")
	ErrInvalidQuantity  = errors.New("quantity must be greater than zero")
	ErrInsufficientStock = errors.New("insufficient stock")
)

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "PENDING"
	OrderStatusPaid      OrderStatus = "PAID"
	OrderStatusShipped   OrderStatus = "SHIPPED"
	OrderStatusCancelled OrderStatus = "CANCELLED"
)

// Order is the aggregate root
type Order struct {
	ID        uuid.UUID
	CustomerID uuid.UUID
	Items     []OrderItem
	Status    OrderStatus
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewOrder creates a new order in pending state
func NewOrder(customerID uuid.UUID, items []OrderItem) (*Order, error) {
	if len(items) == 0 {
		return nil, errors.New("order must have at least one item")
	}

	return &Order{
		ID:        uuid.New(),
		CustomerID: customerID,
		Items:     items,
		Status:    OrderStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

func (o *Order) Total() float64 {
	var total float64
	for _, item := range o.Items {
		total += item.UnitPrice * float64(item.Quantity)
	}
	return total
}

func (o *Order) Pay() error {
	if o.Status != OrderStatusPending {
		return errors.New("order can only be paid when pending")
	}
	o.Status = OrderStatusPaid
	o.UpdatedAt = time.Now()
	return nil
}

type OrderItem struct {
	ProductID uuid.UUID
	Quantity  int
	UnitPrice float64
}
