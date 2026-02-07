package middleware

import (
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
)

func Logging(logger watermill.LoggerAdapter) message.Middleware {
	return func(h message.HandlerFunc) message.HandlerFunc {
		return func(msg *message.Message) ([]*message.Message, error) {
			start := time.Now()
			logger.Info("message_received", watermill.LogFields{
				"message_id": msg.UUID,
				"correlation_id": msg.Metadata.Get("correlation_id"),
			})

			out, err := h(msg)
			if err != nil {
				logger.Error("message_failed", err, watermill.LogFields{
					"message_id": msg.UUID,
					"correlation_id": msg.Metadata.Get("correlation_id"),
					"elapsed": time.Since(start).String(),
				})
				return out, err
			}

			logger.Info("message_processed", watermill.LogFields{
				"message_id": msg.UUID,
				"correlation_id": msg.Metadata.Get("correlation_id"),
				"out_count": len(out),
				"elapsed": time.Since(start).String(),
			})
			return out, nil
		}
	}
}
