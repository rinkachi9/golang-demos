package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/rinkachi/golang-demos/golang-clean-architecture/application"
	"github.com/rinkachi/golang-demos/golang-clean-architecture/domain"
	httphandler "github.com/rinkachi/golang-demos/golang-clean-architecture/infrastructure/http"
	"github.com/rinkachi/golang-demos/golang-clean-architecture/infrastructure/messaging"
	"github.com/rinkachi/golang-demos/golang-clean-architecture/infrastructure/persistence"
)

func main() {
	logger := log.New(os.Stdout, "[CLEAN-ARCH] ", log.LstdFlags)
	watermillLogger := watermill.NewStdLogger(false, false)

	// 1. Infrastructure (Messaging)
	pubSub := gochannel.NewGoChannel(
		gochannel.Config{
			OutputChannelBuffer: 10,
		},
		watermillLogger,
	)

	// Event Bus (Publisher)
	eventBus := messaging.NewWatermillEventBus(pubSub)

	// 2. Infrastructure (Persistence)
	// Retry loop for DB connection
	var db *gorm.DB
	var err error
	dsn := "host=localhost user=user password=password dbname=clean_arch port=5432 sslmode=disable"
	
	for i := 0; i < 10; i++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			break
		}
		logger.Printf("Failed to connect to DB, retrying... (%d/10)", i+1)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		logger.Fatalf("Failed to connect to DB: %v", err)
	}
	logger.Println("Connected to PostgreSQL")

	orderRepo := persistence.NewGormOrderRepository(db)

	// 3. Application
	orderService := application.NewOrderService(orderRepo, eventBus)

	// 4. Infrastructure (Workers / Subscribers)
	router, err := message.NewRouter(message.RouterConfig{}, watermillLogger)
	if err != nil {
		logger.Fatalf("Failed to create Watermill router: %v", err)
	}

	shippingWorker := messaging.NewShippingWorker(logger)
	// Register the worker to listen to OrderPaid events
	// Note: In gochannel, topic name is strict. Our EventBus publishes to "OrderPaid".
	// The Register method uses "OrderPaid" topic.
	shippingWorker.Register(router, pubSub)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		if err := router.Run(ctx); err != nil {
			logger.Fatalf("Watermill router failed: %v", err)
		}
	}()

	// 5. Infrastructure (Transport - HTTP/Gin)
	orderHandler := httphandler.NewOrderHandler(orderService)
	ginRouter := gin.Default()
	orderHandler.RegisterRoutes(ginRouter)

	// 6. Server
	address := ":8080"
	logger.Printf("Starting server on %s", address)
	
	// Graceful shutdown
	srv := &gin.Engine{} // wrapper not needed for Run, but for custom server
	
	go func() {
		if err := ginRouter.Run(address); err != nil {
			logger.Printf("Server stopped: %v", err)
		}
	}()

	// Wait for interrupt signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	logger.Println("Shutting down...")
	cancel() // Stop Watermill router
	// (Add http server shutdown here if using http.Server struct)
	
	time.Sleep(1 * time.Second) // Give workers time to finish
	logger.Println("Exited.")
}
