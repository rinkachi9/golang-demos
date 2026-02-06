package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/gin-gonic/gin"
	"github.com/rinkachi/golang-demos/golang-otel-demo/internal/application"
	appHttp "github.com/rinkachi/golang-demos/golang-otel-demo/internal/infrastructure/http"
	"github.com/rinkachi/golang-demos/golang-otel-demo/internal/infrastructure/messaging"
	"github.com/rinkachi/golang-demos/golang-otel-demo/internal/infrastructure/persistence"
	"github.com/rinkachi/golang-demos/golang-otel-demo/internal/infrastructure/telemetry"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

const (
	ServiceName    = "advanced-otel-demo"
	ServiceVersion = "1.0.0"
	Topic          = "tasks_topic"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// 1. Setup OTel
	otelShutdown, err := telemetry.SetupOTelSDK(ctx, ServiceName, ServiceVersion)
	if err != nil {
		return err
	}
	defer otelShutdown(context.Background())

	// 2. Infrastructure: Persistence
	dsn := "host=localhost user=user password=password dbname=demos port=5432 sslmode=disable"
	db, err := persistence.NewDB(dsn)
	if err != nil {
		return err
	}

	// 3. Infrastructure: Messaging
	brokers := []string{"localhost:9092"}
	publisher, err := messaging.NewPublisher(brokers)
	if err != nil {
		return err
	}
	defer publisher.Close()

	subscriber, err := messaging.NewSubscriber(brokers, "otel-demo-workers")
	if err != nil {
		return err
	}
	defer subscriber.Close()

	// 4. HTTP Handlers
	handler := appHttp.NewHandler(db, publisher, Topic)
	
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(otelgin.Middleware(ServiceName))

	r.POST("/process", handler.HandleProcess)
	r.GET("/status", handler.HandleStatus)

	// 5. Worker
	worker := application.NewWorker(subscriber)
	go func() {
		log.Println("Worker started...")
		if err := worker.Run(ctx, Topic); err != nil {
			log.Printf("Worker did not shut down cleanly: %v", err)
		}
	}()

	// 6. Start Server
	srv := &http.Server{
		Addr:    ":8081",
		Handler: r,
	}

	go func() {
		log.Println("Server listening on :8081")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("listen: %s\n", err)
		}
	}()

	<-ctx.Done()
	return srv.Shutdown(context.Background())
}
