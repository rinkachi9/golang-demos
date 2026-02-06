package application

import (
	"context"
	"log"
	"net/http"

	"github.com/ThreeDotsLabs/watermill/message"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

var tracer = otel.Tracer("otel-demo-worker")

type Worker struct {
	subscriber message.Subscriber
}

func NewWorker(subscriber message.Subscriber) *Worker {
	return &Worker{subscriber: subscriber}
}

func (w *Worker) Run(ctx context.Context, topic string) error {
	messages, err := w.subscriber.Subscribe(ctx, topic)
	if err != nil {
		return err
	}

	// Blocking process loop
	w.processMessages(messages)
	return nil
}

func (w *Worker) processMessages(messages <-chan *message.Message) {
	// Create a custom client with OTel transport
	httpClient := http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
	}

	for msg := range messages {
		// Extract trace context
		ctx := otel.GetTextMapPropagator().Extract(context.Background(), propagation.MapCarrier(msg.Metadata))
		
		// Start a new span linked to the producer
		ctx, span := tracer.Start(ctx, "process_message", trace.WithAttributes(
			attribute.String("message.id", msg.UUID),
		))

		log.Printf("Processing message: %s", string(msg.Payload))

		// Call Internal API (Traced)
		req, _ := http.NewRequestWithContext(ctx, "GET", "http://localhost:8081/status", nil)
		resp, err := httpClient.Do(req)
		if err != nil {
			span.RecordError(err)
			log.Printf("Callback failed: %v", err)
		} else {
			resp.Body.Close()
		}
		
		span.End()
		msg.Ack()
	}
}
