package messaging

import (
	"os"

	"github.com/Shopify/sarama"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-kafka/v3/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill/message"
)

func NewKafkaPublisher(brokers []string, logger watermill.LoggerAdapter) message.Publisher {
	cfg := kafka.DefaultSaramaSyncPublisherConfig()
	cfg.Producer.Return.Successes = true

	publisher, err := kafka.NewPublisher(kafka.PublisherConfig{
		Brokers:               brokers,
		Marshaler:             kafka.DefaultMarshaler{},
		OverwriteSaramaConfig: cfg,
		Tracer:                kafka.NewOTELSaramaTracer(),
	}, logger)
	if err != nil {
		logger.Error("kafka_publisher_failed", err, nil)
		os.Exit(1)
	}
	return publisher
}

func NewKafkaSubscriber(brokers []string, group string, logger watermill.LoggerAdapter) message.Subscriber {
	cfg := kafka.DefaultSaramaSubscriberConfig()
	cfg.Consumer.Offsets.Initial = sarama.OffsetOldest

	subscriber, err := kafka.NewSubscriber(kafka.SubscriberConfig{
		Brokers:               brokers,
		Unmarshaler:           kafka.DefaultMarshaler{},
		ConsumerGroup:         group,
		InitializeTopicDetails: &sarama.TopicDetail{
			NumPartitions:     3,
			ReplicationFactor: 1,
		},
		OverwriteSaramaConfig: cfg,
		Tracer:                kafka.NewOTELSaramaTracer(),
	}, logger)
	if err != nil {
		logger.Error("kafka_subscriber_failed", err, nil)
		os.Exit(1)
	}
	return subscriber
}
