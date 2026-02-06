package messaging

import (
	"context"
	"encoding/json"
	"log"
	
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/rinkachi/golang-demos/golang-clean-architecture/domain"
)

type ShippingWorker struct {
	logger *log.Logger
}

func NewShippingWorker(logger *log.Logger) *ShippingWorker {
	return &ShippingWorker{
		logger: logger,
	}
}

func (w *ShippingWorker) HandleOrderPaid(msg *message.Message) error {
	var event domain.OrderPaid
	if err := json.Unmarshal(msg.Payload, &event); err != nil {
		return err // In real app, maybe send to DLQ
	}

	w.logger.Printf("[SHIPPING] Processing shipment for Order %s. Amount paid: %.2f", event.OrderID, event.TotalAmount)
	
	// Simulate shipping processing
	// ... 
	
	return nil
}

// RegisterHandlers registers the worker methods to the Watermill router
func (w *ShippingWorker) Register(router *message.Router, subscriber message.Subscriber) {
	router.AddNoPublisherHandler(
		"shipping_order_paid_handler",
		"OrderPaid", // handler name
		subscriber,
		w.HandleOrderPaid,
	)
}
