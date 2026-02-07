package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/components/metrics"
	"github.com/ThreeDotsLabs/watermill/message"
	wmiddleware "github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/ThreeDotsLabs/watermill/message/router/plugin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"

	"github.com/rinkachi/golang-demos/golang-watermill-demo/internal/application/usecase"
	"github.com/rinkachi/golang-demos/golang-watermill-demo/internal/config"
	"github.com/rinkachi/golang-demos/golang-watermill-demo/internal/domain/topics"
	"github.com/rinkachi/golang-demos/golang-watermill-demo/internal/infra/messaging"
	infraMiddleware "github.com/rinkachi/golang-demos/golang-watermill-demo/internal/infra/messaging/middleware"
	"github.com/rinkachi/golang-demos/golang-watermill-demo/internal/infra/telemetry"
)

func main() {
	cfg := config.LoadWorker()
	logger := watermill.NewStdLogger(true, false)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if shutdown, err := telemetry.Setup(ctx, cfg.ServiceName, cfg.ServiceVersion, cfg.OtelEndpoint); err != nil {
		logger.Error("otel_setup_failed", err, nil)
	} else {
		defer shutdown(ctx)
	}

	promRegistry, closeMetrics := metrics.CreateRegistryAndServeHTTP(cfg.MetricsAddr)
	defer closeMetrics()
	metricsBuilder := metrics.NewPrometheusMetricsBuilder(promRegistry, "watermill", "worker")

	kafkaPublisher := messaging.NewKafkaPublisher(cfg.KafkaBrokers, logger)
	kafkaSubscriber := messaging.NewKafkaSubscriber(cfg.KafkaBrokers, "watermill-worker", logger)
	rabbitPublisher := messaging.NewRabbitPublisher(cfg.RabbitURL, logger)
	rabbitSubscriber := messaging.NewRabbitSubscriber(cfg.RabbitURL, "watermill-worker", messaging.DeadLetterConfig{
		Exchange:   "watermill.dlx",
		Queue:      "watermill.dead_letter",
		RoutingKey: "dead-letter",
	}, logger)
	defer kafkaPublisher.Close()
	defer kafkaSubscriber.Close()
	defer rabbitPublisher.Close()
	defer rabbitSubscriber.Close()

	router, err := message.NewRouter(message.RouterConfig{CloseTimeout: 10 * time.Second}, logger)
	if err != nil {
		logger.Error("router_create_failed", err, nil)
		return
	}

	router.AddMiddleware(
		wmiddleware.Recoverer,
		wmiddleware.CorrelationID,
		wmiddleware.Timeout(5*time.Second),
		wmiddleware.Retry{
			MaxRetries:      3,
			InitialInterval: 100 * time.Millisecond,
			Logger:          logger,
		}.Middleware,
		wmiddleware.PoisonQueue{
			MaxRetries:              3,
			FailedMessagesPublisher: rabbitPublisher,
			FailedMessageTopic:      topics.PoisonQueue,
			Logger:                  logger,
		}.Middleware,
		infraMiddleware.Logging(logger),
		infraMiddleware.Tracing(otel.Tracer(cfg.ServiceName), propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		)),
	)

	metricsBuilder.AddPrometheusRouterMetrics(router)
	router.AddPlugin(plugin.SignalsHandler)

	router.AddHandler(
		"orders_enricher",
		topics.OrdersIncoming,
		kafkaSubscriber,
		topics.OrdersEnriched,
		kafkaPublisher,
		withHandlerMiddleware(func(msg *message.Message) ([]*message.Message, error) {
			order, err := usecase.DecodeOrder(msg.Payload)
			if err != nil {
				return nil, err
			}

			enriched := usecase.EnrichOrder(order)
			payload, err := usecase.EncodeJSON(enriched)
			if err != nil {
				return nil, err
			}

			out := message.NewMessage(watermill.NewUUID(), payload)
			out.Metadata = copyMetadata(msg.Metadata)
			out.Metadata.Set("event_type", "order.enriched")
			out.SetContext(msg.Context())
			return []*message.Message{out}, nil
		}, wmiddleware.Retry{
			MaxRetries:      5,
			InitialInterval: 200 * time.Millisecond,
			MaxInterval:     2 * time.Second,
			Multiplier:      2,
			Logger:          logger,
		}.Middleware),
	)

	router.AddHandler(
		"orders_notify",
		topics.OrdersEnriched,
		kafkaSubscriber,
		topics.Notifications,
		rabbitPublisher,
		withHandlerMiddleware(func(msg *message.Message) ([]*message.Message, error) {
			enriched, err := usecase.DecodeEnriched(msg.Payload)
			if err != nil {
				return nil, err
			}

			notification := usecase.BuildNotification(enriched)
			payload, err := usecase.EncodeJSON(notification)
			if err != nil {
				return nil, err
			}

			out := message.NewMessage(watermill.NewUUID(), payload)
			out.Metadata = copyMetadata(msg.Metadata)
			out.Metadata.Set("event_type", "notification.email")
			out.SetContext(msg.Context())
			return []*message.Message{out}, nil
		}, wmiddleware.Retry{
			MaxRetries:      2,
			InitialInterval: 500 * time.Millisecond,
			MaxInterval:     3 * time.Second,
			Multiplier:      2,
			Logger:          logger,
		}.Middleware),
	)

	router.AddHandler(
		"orders_realtime",
		topics.OrdersEnriched,
		kafkaSubscriber,
		topics.RealtimeUpdates,
		kafkaPublisher,
		withHandlerMiddleware(func(msg *message.Message) ([]*message.Message, error) {
			enriched, err := usecase.DecodeEnriched(msg.Payload)
			if err != nil {
				return nil, err
			}

			update := usecase.BuildRealtimeUpdate(enriched)
			payload, err := usecase.EncodeJSON(update)
			if err != nil {
				return nil, err
			}

			out := message.NewMessage(watermill.NewUUID(), payload)
			out.Metadata = copyMetadata(msg.Metadata)
			out.Metadata.Set("event_type", "realtime.order")
			out.SetContext(msg.Context())
			return []*message.Message{out}, nil
		}, wmiddleware.Retry{
			MaxRetries:      1,
			InitialInterval: 100 * time.Millisecond,
			MaxInterval:     1 * time.Second,
			Multiplier:      2,
			Logger:          logger,
		}.Middleware),
	)

	router.AddNoPublisherHandler(
		"orders_audit_consumer",
		topics.OrdersAudit,
		rabbitSubscriber,
		withNoPublisherHandlerMiddleware(func(msg *message.Message) error {
			logger.Info("order_audit_received", watermill.LogFields{
				"payload": string(msg.Payload),
			})
			return nil
		}, wmiddleware.Retry{
			MaxRetries:      3,
			InitialInterval: 250 * time.Millisecond,
			MaxInterval:     2 * time.Second,
			Multiplier:      2,
			Logger:          logger,
		}.Middleware),
	)

	router.AddNoPublisherHandler(
		"poison_queue_logger",
		topics.PoisonQueue,
		rabbitSubscriber,
		withNoPublisherHandlerMiddleware(func(msg *message.Message) error {
			logger.Error("poison_message", fmt.Errorf("moved_to_poison_queue"), watermill.LogFields{
				"payload":  string(msg.Payload),
				"metadata": msg.Metadata,
			})
			return nil
		}, wmiddleware.Retry{
			MaxRetries:      1,
			InitialInterval: 200 * time.Millisecond,
			MaxInterval:     1 * time.Second,
			Multiplier:      2,
			Logger:          logger,
		}.Middleware),
	)

	if err := router.Run(ctx); err != nil {
		logger.Error("router_run_failed", err, nil)
		os.Exit(1)
	}
}

func copyMetadata(in message.Metadata) message.Metadata {
	out := make(message.Metadata, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func withHandlerMiddleware(h message.HandlerFunc, m ...message.HandlerMiddleware) message.HandlerFunc {
	for i := len(m) - 1; i >= 0; i-- {
		h = m[i](h)
	}
	return h
}

func withNoPublisherHandlerMiddleware(h message.NoPublisherHandlerFunc, m ...message.HandlerMiddleware) message.NoPublisherHandlerFunc {
	wrapped := withHandlerMiddleware(func(msg *message.Message) ([]*message.Message, error) {
		return nil, h(msg)
	}, m...)

	return func(msg *message.Message) error {
		_, err := wrapped(msg)
		return err
	}
}
