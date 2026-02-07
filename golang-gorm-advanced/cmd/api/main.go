package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/rinkachi/golang-demos/golang-gorm-advanced/internal/application"
	"github.com/rinkachi/golang-demos/golang-gorm-advanced/internal/config"
	"github.com/rinkachi/golang-demos/golang-gorm-advanced/internal/infra/db"
	"github.com/rinkachi/golang-demos/golang-gorm-advanced/internal/infra/metrics"
	"github.com/rinkachi/golang-demos/golang-gorm-advanced/internal/infra/persistence"
	httpapi "github.com/rinkachi/golang-demos/golang-gorm-advanced/internal/interfaces/http"
)

func main() {
	cfg := config.LoadAPI()

	database, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db connect failed: %v", err)
	}

	userRepo := persistence.NewUserRepository(database)
	orderRepo := persistence.NewOrderRepository(database)
	service := application.NewService(userRepo, orderRepo)

	metricsServer := metrics.StartServer(cfg.MetricsAddr)
	metricsRegistry := metrics.New()

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(metricsRegistry.Middleware())

	handlers := httpapi.NewHandlers(service)
	handlers.Register(router)

	srv := &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: router,
	}

	go func() {
		log.Printf("GORM API listening on %s", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http server failed: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
	_ = metricsServer.Shutdown(ctx)
}
