package domain

import "time"

// ProcessLog represents a log of a processed request
type ProcessLog struct {
	ID        uint `gorm:"primaryKey"`
	RequestID string
	Payload   string
	CreatedAt time.Time
}
