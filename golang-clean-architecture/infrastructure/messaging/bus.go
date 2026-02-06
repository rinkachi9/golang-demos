package messaging

import (
	"encoding/json"
	
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/rinkachi/golang-demos/golang-clean-architecture/domain"
)

type WatermillEventBus struct {
	publisher message.Publisher
}

func NewWatermillEventBus(publisher message.Publisher) *WatermillEventBus {
	return &WatermillEventBus{
		publisher: publisher,
	}
}

func (b *WatermillEventBus) Publish(event domain.Event) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	msg := message.NewMessage(watermill.NewUUID(), payload)
	
	// Use the event name as the topic
	return b.publisher.Publish(event.EventName(), msg)
}
