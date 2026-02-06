package persistence

import (
	"context"
	"encoding/json"
	
	"github.com/google/uuid"
	"gorm.io/gorm"
	
	"github.com/rinkachi/golang-demos/golang-clean-architecture/domain"
)

// GormOrder is the DB model for Order
type GormOrder struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key"`
	CustomerID uuid.UUID `gorm:"type:uuid"`
	Status     string
	Total      float64
	ItemsJSON  []byte     `gorm:"type:jsonb"` // Simple storage for items for this demo
	CreatedAt  string // Store as string or time.Time. GORM handles time.Time automatically usually, but let's stick to simple mapping.
	// Actually GORM handles time.Time well.
}

// ToDomain maps DB model to Domain entity
func (g *GormOrder) ToDomain() (*domain.Order, error) {
	var items []domain.OrderItem
	if err := json.Unmarshal(g.ItemsJSON, &items); err != nil {
		return nil, err
	}
	
	// Reconstruct aggregate
	// Note: In real app, we might map fields more carefully
	return &domain.Order{
		ID:         g.ID,
		CustomerID: g.CustomerID,
		Items:      items,
		Status:     domain.OrderStatus(g.Status),
		// CreatedAt: ...
	}, nil
}

type GormOrderRepository struct {
	db *gorm.DB
}

func NewGormOrderRepository(db *gorm.DB) *GormOrderRepository {
	// AutoMigrate
	db.AutoMigrate(&GormOrder{})
	return &GormOrderRepository{db: db}
}

func (r *GormOrderRepository) Save(ctx context.Context, order *domain.Order) error {
	itemsJSON, err := json.Marshal(order.Items)
	if err != nil {
		return err
	}

	model := GormOrder{
		ID:         order.ID,
		CustomerID: order.CustomerID,
		Status:     string(order.Status),
		Total:      order.Total(),
		ItemsJSON:  itemsJSON,
	}

	return r.db.WithContext(ctx).Save(&model).Error
}

func (r *GormOrderRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	var model GormOrder
	if err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrOrderNotFound
		}
		return nil, err
	}

	return model.ToDomain()
}

func (r *GormOrderRepository) FindAll(ctx context.Context) ([]*domain.Order, error) {
	var models []GormOrder
	if err := r.db.WithContext(ctx).Find(&models).Error; err != nil {
		return nil, err
	}

	orders := make([]*domain.Order, len(models))
	for i, m := range models {
		o, err := m.ToDomain()
		if err != nil {
			return nil, err
		}
		orders[i] = o
	}
	return orders, nil
}
