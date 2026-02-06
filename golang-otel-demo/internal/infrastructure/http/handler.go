package http

import (
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/gin-gonic/gin"
	"github.com/rinkachi/golang-demos/golang-otel-demo/internal/domain"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

type Handler struct {
	db        *gorm.DB
	publisher message.Publisher
	topic     string
}

func NewHandler(db *gorm.DB, publisher message.Publisher, topic string) *Handler {
	return &Handler{
		db:        db,
		publisher: publisher,
		topic:     topic,
	}
}

func (h *Handler) HandleProcess(c *gin.Context) {
	span := trace.SpanFromContext(c.Request.Context())
	span.AddEvent("Received process request")

	var req struct {
		Data string `json:"data"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "bad request"})
		return
	}

	// A. DB Operation
	logEntry := domain.ProcessLog{
		RequestID: span.SpanContext().TraceID().String(),
		Payload:   req.Data,
	}
	if err := h.db.WithContext(c.Request.Context()).Create(&logEntry).Error; err != nil {
		span.RecordError(err)
		c.JSON(500, gin.H{"error": "db error"})
		return
	}

	// B. Publish to Kafka
	msg := message.NewMessage(watermill.NewUUID(), []byte(req.Data))
	
	// Inject trace context
	otel.GetTextMapPropagator().Inject(c.Request.Context(), propagation.MapCarrier(msg.Metadata))

	if err := h.publisher.Publish(h.topic, msg); err != nil {
		span.RecordError(err)
		c.JSON(500, gin.H{"error": "kafka error"})
		return
	}

	c.JSON(202, gin.H{"status": "accepted", "trace_id": span.SpanContext().TraceID().String()})
}

func (h *Handler) HandleStatus(c *gin.Context) {
	span := trace.SpanFromContext(c.Request.Context())
	span.AddEvent("Internal status check")
	c.JSON(200, gin.H{"status": "ok"})
}
