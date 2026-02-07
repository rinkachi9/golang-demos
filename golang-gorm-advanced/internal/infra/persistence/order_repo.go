package persistence

import (
	"context"

	"github.com/rinkachi/golang-demos/golang-gorm-advanced/internal/domain/model"
	"github.com/rinkachi/golang-demos/golang-gorm-advanced/internal/domain/scopes"
	"gorm.io/gorm"
)

type OrderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) CreateOrderWithItems(ctx context.Context, order *model.Order, items []model.OrderItem, audit model.AuditLog) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(order).Error; err != nil {
			return err
		}

		for i := range items {
			items[i].OrderID = order.ID
		}

		if len(items) > 0 {
			if err := tx.Create(&items).Error; err != nil {
				return err
			}
		}

		audit.EntityID = order.ID
		return tx.Create(&audit).Error
	})
}

func (r *OrderRepository) GetByID(ctx context.Context, id uint) (*model.Order, error) {
	var order model.Order
	if err := r.db.WithContext(ctx).
		Preload("Items").
		Preload("User").
		First(&order, id).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *OrderRepository) List(ctx context.Context, minTotal float64, recentDays int) ([]model.Order, error) {
	query := r.db.WithContext(ctx).Model(&model.Order{}).
		Preload("Items").
		Preload("User").
		Scopes(scopes.OrdersMinTotal(minTotal), scopes.OrdersRecentDays(recentDays)).
		Order("created_at desc")

	var orders []model.Order
	if err := query.Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}
