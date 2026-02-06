package messaging

import (
	"fmt"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-kafka/v2/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill/message"
)

func NewPublisher(brokers []string) (message.Publisher, error) {
	logger := watermill.NewStdLogger(false, false)
	return kafka.NewPublisher(
		kafka.PublisherConfig{
			Brokers:   brokers,
			Marshaler: kafka.DefaultMarshaler{},
		},
		logger,
	)
}

func NewSubscriber(brokers []string, consumerGroup string) (message.Subscriber, error) {
	logger := watermill.NewStdLogger(false, false)
	return kafka.NewSubscriber(
		kafka.SubscriberConfig{
			Brokers:       brokers,
			Unmarshaler:   kafka.DefaultMarshaler{},
			ConsumerGroup: consumerGroup,
		},
		logger,
	)
}
