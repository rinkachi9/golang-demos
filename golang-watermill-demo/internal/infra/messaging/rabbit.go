package messaging

import (
	"os"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-amqp/v3/pkg/amqp"
	"github.com/ThreeDotsLabs/watermill/message"
	amqp091 "github.com/rabbitmq/amqp091-go"
)

func NewRabbitPublisher(url string, logger watermill.LoggerAdapter) message.Publisher {
	cfg := amqp.NewDurablePubSubConfig(url, "watermill-demo")
	publisher, err := amqp.NewPublisher(cfg, logger)
	if err != nil {
		logger.Error("rabbit_publisher_failed", err, nil)
		os.Exit(1)
	}
	return publisher
}

type DeadLetterConfig struct {
	Exchange   string
	Queue      string
	RoutingKey string
}

func NewRabbitSubscriber(url, consumerGroup string, dlx DeadLetterConfig, logger watermill.LoggerAdapter) message.Subscriber {
	cfg := amqp.NewDurablePubSubConfig(url, amqp.GenerateQueueNameTopicNameWithSuffix(consumerGroup))
	cfg.Queue.Arguments = amqp091.Table{
		"x-dead-letter-exchange":    dlx.Exchange,
		"x-dead-letter-routing-key": dlx.RoutingKey,
	}
	cfg.TopologyBuilder = DeadLetterTopology{
		DLX: dlx,
	}

	subscriber, err := amqp.NewSubscriber(cfg, logger)
	if err != nil {
		logger.Error("rabbit_subscriber_failed", err, nil)
		os.Exit(1)
	}
	return subscriber
}

type DeadLetterTopology struct {
	DLX DeadLetterConfig
	amqp.DefaultTopologyBuilder
}

func (t DeadLetterTopology) BuildTopology(channel *amqp091.Channel, queueName string, exchangeName string, config amqp.Config, logger watermill.LoggerAdapter) error {
	if err := t.DefaultTopologyBuilder.BuildTopology(channel, queueName, exchangeName, config, logger); err != nil {
		return err
	}

	if err := channel.ExchangeDeclare(
		t.DLX.Exchange,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return err
	}

	dlq, err := channel.QueueDeclare(
		t.DLX.Queue,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	return channel.QueueBind(
		dlq.Name,
		t.DLX.RoutingKey,
		t.DLX.Exchange,
		false,
		nil,
	)
}
