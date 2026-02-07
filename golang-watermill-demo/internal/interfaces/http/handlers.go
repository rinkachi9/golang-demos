package httpapi

import (
	"net/http"
	"strings"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/gin-gonic/gin"

	"github.com/rinkachi/golang-demos/golang-watermill-demo/internal/application/usecase"
	"github.com/rinkachi/golang-demos/golang-watermill-demo/internal/domain/model"
	"github.com/rinkachi/golang-demos/golang-watermill-demo/internal/domain/topics"
)

type Publisher interface {
	Publish(topic string, messages ...*message.Message) error
}

type Handlers struct {
	logger          watermill.LoggerAdapter
	kafkaPublisher  Publisher
	rabbitPublisher Publisher
}

func NewHandlers(logger watermill.LoggerAdapter, kafkaPub Publisher, rabbitPub Publisher) *Handlers {
	return &Handlers{
		logger:          logger,
		kafkaPublisher:  kafkaPub,
		rabbitPublisher: rabbitPub,
	}
}

func (h *Handlers) Register(r *gin.Engine) {
	r.POST("/api/order", h.publishOrder)
	r.GET("/healthz", h.health)
}

func (h *Handlers) health(c *gin.Context) {
	c.String(http.StatusOK, "ok")
}

func (h *Handlers) publishOrder(c *gin.Context) {
	var order model.Order
	if err := c.ShouldBindJSON(&order); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	if order.ID == "" {
		order.ID = watermill.NewUUID()
	}
	if order.CreatedAt.IsZero() {
		order.CreatedAt = time.Now()
	}

	payload, err := usecase.EncodeJSON(order)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "marshal error"})
		return
	}

	meta := message.Metadata{}
	meta.Set("correlation_id", correlationID(c.Request))
	meta.Set("event_type", "order.created")

	msg := message.NewMessage(watermill.NewUUID(), payload)
	msg.Metadata = meta
	msg.SetContext(c.Request.Context())

	if err := h.kafkaPublisher.Publish(topics.OrdersIncoming, msg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "kafka publish failed"})
		return
	}

	if err := h.rabbitPublisher.Publish(topics.OrdersAudit, msg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "rabbit publish failed"})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"status": "queued", "order_id": order.ID})
}

func correlationID(r *http.Request) string {
	candidate := strings.TrimSpace(r.Header.Get("X-Correlation-ID"))
	if candidate != "" {
		return candidate
	}
	return watermill.NewUUID()
}
