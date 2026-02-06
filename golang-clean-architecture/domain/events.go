package domain

import (
	"time"

	"github.com/google/uuid"
)

// Event is a marker interface for domain events
type Event interface {
	EventName() string
}

type OrderPaid struct {
	OrderID   uuid.UUID
	PaidAt    time.Time
	TotalAmount float64
}

func (e OrderPaid) EventName() string {
	return "OrderPaid"
}

// EventBus defines how application events are published
type EventBus interface {
	Publish(event Event) error
}
