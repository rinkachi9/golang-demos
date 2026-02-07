package middleware

import (
	"github.com/ThreeDotsLabs/watermill/message"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

func Tracing(tracer trace.Tracer, propagator propagation.TextMapPropagator) message.Middleware {
	return func(h message.HandlerFunc) message.HandlerFunc {
		return func(msg *message.Message) ([]*message.Message, error) {
			ctx := propagator.Extract(msg.Context(), propagation.MapCarrier(msg.Metadata))
			ctx, span := tracer.Start(ctx, "watermill.handle",
				trace.WithSpanKind(trace.SpanKindConsumer),
				trace.WithAttributes(
					attribute.String("message.id", msg.UUID),
					attribute.String("correlation.id", msg.Metadata.Get("correlation_id")),
				),
			)
			defer span.End()

			msg.SetContext(ctx)

			out, err := h(msg)
			if err != nil {
				span.RecordError(err)
			}

			for _, outMsg := range out {
				propagator.Inject(ctx, propagation.MapCarrier(outMsg.Metadata))
			}

			return out, err
		}
	}
}
