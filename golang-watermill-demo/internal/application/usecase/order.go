package usecase

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/rinkachi/golang-demos/golang-watermill-demo/internal/domain/model"
)

func EnrichOrder(order model.Order) model.EnrichedOrder {
	return model.EnrichedOrder{
		Order:      order,
		RiskScore:  rand.Intn(100),
		EnrichedAt: time.Now(),
	}
}

func BuildNotification(enriched model.EnrichedOrder) model.Notification {
	return model.Notification{
		OrderID: enriched.ID,
		Channel: "email",
		Message: fmt.Sprintf("Order %s for %s is ready (risk=%d)", enriched.ID, enriched.Customer, enriched.RiskScore),
	}
}

func BuildRealtimeUpdate(enriched model.EnrichedOrder) model.RealtimeUpdate {
	return model.RealtimeUpdate{
		Type:      "order.enriched",
		Timestamp: time.Now(),
		Payload:   enriched,
	}
}

func EncodeJSON(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func DecodeOrder(payload []byte) (model.Order, error) {
	var order model.Order
	return order, json.Unmarshal(payload, &order)
}

func DecodeEnriched(payload []byte) (model.EnrichedOrder, error) {
	var enriched model.EnrichedOrder
	return enriched, json.Unmarshal(payload, &enriched)
}
