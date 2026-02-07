package model

import "time"

type Order struct {
	ID        string    `json:"id"`
	Customer  string    `json:"customer"`
	Total     float64   `json:"total"`
	CreatedAt time.Time `json:"created_at"`
}

type EnrichedOrder struct {
	Order
	RiskScore  int       `json:"risk_score"`
	EnrichedAt time.Time `json:"enriched_at"`
}

type Notification struct {
	OrderID string `json:"order_id"`
	Channel string `json:"channel"`
	Message string `json:"message"`
}

type RealtimeUpdate struct {
	Type      string      `json:"type"`
	Timestamp time.Time   `json:"timestamp"`
	Payload   interface{} `json:"payload"`
}
