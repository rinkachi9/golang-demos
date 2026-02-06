package application

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/rinkachi/golang-demos/golang-clean-architecture/domain"
)

type CreateOrderInput struct {
	CustomerID uuid.UUID
	Items      []CreateOrderItemInput
}

type CreateOrderItemInput struct {
	ProductID uuid.UUID
	Quantity  int
	UnitPrice float64
}

type OrderOutput struct {
	ID         uuid.UUID
	CustomerID uuid.UUID
	Status     string
	Total      float64
	CreatedAt  time.Time
}

type OrderService struct {
	repo     domain.OrderRepository
	eventBus domain.EventBus
}

func NewOrderService(repo domain.OrderRepository, eventBus domain.EventBus) *OrderService {
	return &OrderService{
		repo:     repo,
		eventBus: eventBus,
	}
}

func (s *OrderService) CreateOrder(ctx context.Context, input CreateOrderInput) (*OrderOutput, error) {
	var items []domain.OrderItem
	for _, i := range input.Items {
		items = append(items, domain.OrderItem{
			ProductID: i.ProductID,
			Quantity:  i.Quantity,
			UnitPrice: i.UnitPrice,
		})
	}

	order, err := domain.NewOrder(input.CustomerID, items)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Save(ctx, order); err != nil {
		return nil, err
	}

	return s.toOutput(order), nil
}

func (s *OrderService) PayOrder(ctx context.Context, orderID uuid.UUID) error {
	order, err := s.repo.FindByID(ctx, orderID)
	if err != nil {
		return err
	}

	if err := order.Pay(); err != nil {
		return err
	}

	if err := s.repo.Save(ctx, order); err != nil {
		return err
	}

	// Publish domain event
	event := domain.OrderPaid{
		OrderID:     order.ID,
		PaidAt:      time.Now(),
		TotalAmount: order.Total(),
	}
	
	// In a real transactional outbox pattern, this would be part of the transaction.
	// For this demo, we publish directly.
	return s.eventBus.Publish(event)
}

func (s *OrderService) GetOrder(ctx context.Context, id uuid.UUID) (*OrderOutput, error) {
	order, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.toOutput(order), nil
}

func (s *OrderService) toOutput(order *domain.Order) *OrderOutput {
	return &OrderOutput{
		ID:         order.ID,
		CustomerID: order.CustomerID,
		Status:     string(order.Status),
		Total:      order.Total(),
		CreatedAt:  order.CreatedAt,
	}
}
