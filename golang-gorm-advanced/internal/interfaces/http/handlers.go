package httpapi

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/rinkachi/golang-demos/golang-gorm-advanced/internal/application"
	"github.com/rinkachi/golang-demos/golang-gorm-advanced/internal/domain/model"
)

type Handlers struct {
	service *application.Service
}

func NewHandlers(service *application.Service) *Handlers {
	return &Handlers{service: service}
}

func (h *Handlers) Register(r *gin.Engine) {
	r.GET("/healthz", h.health)

	r.POST("/users", h.createUser)
	r.GET("/users", h.listUsers)
	r.GET("/users/:id", h.getUser)
	r.PATCH("/users/:id/deactivate", h.deactivateUser)

	r.POST("/orders", h.createOrder)
	r.GET("/orders", h.listOrders)
	r.GET("/orders/:id", h.getOrder)
}

func (h *Handlers) health(c *gin.Context) {
	c.String(http.StatusOK, "ok")
}

func (h *Handlers) createUser(c *gin.Context) {
	var req struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	user := &model.User{
		Email:  req.Email,
		Name:   req.Name,
		Active: true,
	}
	if err := h.service.CreateUser(c.Request.Context(), user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, user)
}

func (h *Handlers) listUsers(c *gin.Context) {
	activeOnly := c.Query("active") == "true"
	domain := c.Query("domain")

	users, err := h.service.ListUsers(c.Request.Context(), activeOnly, domain)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, users)
}

func (h *Handlers) getUser(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	user, err := h.service.GetUser(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *Handlers) deactivateUser(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	user, err := h.service.DeactivateUser(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *Handlers) createOrder(c *gin.Context) {
	var req struct {
		UserID uint `json:"user_id"`
		Status string `json:"status"`
		Items  []struct {
			SKU   string  `json:"sku"`
			Qty   int     `json:"qty"`
			Price float64 `json:"price"`
		} `json:"items"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.UserID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	var total float64
	items := make([]model.OrderItem, 0, len(req.Items))
	for _, item := range req.Items {
		items = append(items, model.OrderItem{
			SKU:   item.SKU,
			Qty:   item.Qty,
			Price: item.Price,
		})
		total += item.Price * float64(item.Qty)
	}

	order := &model.Order{
		UserID: req.UserID,
		Status: req.Status,
		Total:  total,
	}

	if err := h.service.CreateOrderTx(c.Request.Context(), order, items); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, order)
}

func (h *Handlers) getOrder(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	order, err := h.service.GetOrder(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, order)
}

func (h *Handlers) listOrders(c *gin.Context) {
	minTotal, _ := strconv.ParseFloat(c.Query("min_total"), 64)
	recentDays, _ := strconv.Atoi(c.Query("recent_days"))

	orders, err := h.service.ListOrders(c.Request.Context(), minTotal, recentDays)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, orders)
}

func parseID(raw string) (uint, error) {
	id64, err := strconv.ParseUint(raw, 10, 64)
	return uint(id64), err
}
