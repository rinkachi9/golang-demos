package model

import "time"

type User struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Email     string         `gorm:"uniqueIndex;size:255" json:"email"`
	Name      string         `gorm:"size:120" json:"name"`
	Active    bool           `gorm:"default:true" json:"active"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt *time.Time     `gorm:"index" json:"deleted_at,omitempty"`
	Orders    []Order        `json:"orders,omitempty"`
}

type Order struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	UserID    uint       `gorm:"index" json:"user_id"`
	User      User       `json:"user,omitempty"`
	Status    string     `gorm:"size:40;index" json:"status"`
	Total     float64    `gorm:"type:numeric(12,2)" json:"total"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	Items     []OrderItem `json:"items,omitempty"`
}

type OrderItem struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	OrderID   uint      `gorm:"index" json:"order_id"`
	SKU       string    `gorm:"size:64" json:"sku"`
	Qty       int       `json:"qty"`
	Price     float64   `gorm:"type:numeric(12,2)" json:"price"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AuditLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Entity    string    `gorm:"size:40;index" json:"entity"`
	EntityID  uint      `gorm:"index" json:"entity_id"`
	Action    string    `gorm:"size:80" json:"action"`
	Details   string    `gorm:"type:text" json:"details"`
	CreatedAt time.Time `json:"created_at"`
}
