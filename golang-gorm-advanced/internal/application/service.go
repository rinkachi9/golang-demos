package application

import (
	"context"
	"fmt"

	"github.com/rinkachi/golang-demos/golang-gorm-advanced/internal/domain/model"
)

type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	GetByID(ctx context.Context, id uint) (*model.User, error)
	List(ctx context.Context, activeOnly bool, domain string) ([]model.User, error)
	Update(ctx context.Context, user *model.User) error
}

type OrderRepository interface {
	CreateOrderWithItems(ctx context.Context, order *model.Order, items []model.OrderItem, audit model.AuditLog) error
	GetByID(ctx context.Context, id uint) (*model.Order, error)
	List(ctx context.Context, minTotal float64, recentDays int) ([]model.Order, error)
}

type Service struct {
	users  UserRepository
	orders OrderRepository
}

func NewService(users UserRepository, orders OrderRepository) *Service {
	return &Service{users: users, orders: orders}
}

func (s *Service) CreateUser(ctx context.Context, user *model.User) error {
	return s.users.Create(ctx, user)
}

func (s *Service) GetUser(ctx context.Context, id uint) (*model.User, error) {
	return s.users.GetByID(ctx, id)
}

func (s *Service) ListUsers(ctx context.Context, activeOnly bool, domain string) ([]model.User, error) {
	return s.users.List(ctx, activeOnly, domain)
}

func (s *Service) DeactivateUser(ctx context.Context, id uint) (*model.User, error) {
	user, err := s.users.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	user.Active = false
	if err := s.users.Update(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *Service) CreateOrderTx(ctx context.Context, order *model.Order, items []model.OrderItem) error {
	audit := model.AuditLog{
		Entity:   "order",
		EntityID: order.ID,
		Action:   "create",
		Details:  fmt.Sprintf("order for user %d, items=%d", order.UserID, len(items)),
	}
	return s.orders.CreateOrderWithItems(ctx, order, items, audit)
}

func (s *Service) GetOrder(ctx context.Context, id uint) (*model.Order, error) {
	return s.orders.GetByID(ctx, id)
}

func (s *Service) ListOrders(ctx context.Context, minTotal float64, recentDays int) ([]model.Order, error) {
	return s.orders.List(ctx, minTotal, recentDays)
}
