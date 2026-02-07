package main

import (
	"context"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/components/metrics"
	"github.com/ThreeDotsLabs/watermill/message"
	wmiddleware "github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/ThreeDotsLabs/watermill/message/router/plugin"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"

	"github.com/rinkachi/golang-demos/golang-watermill-demo/internal/config"
	"github.com/rinkachi/golang-demos/golang-watermill-demo/internal/domain/topics"
	"github.com/rinkachi/golang-demos/golang-watermill-demo/internal/infra/messaging"
	infraMiddleware "github.com/rinkachi/golang-demos/golang-watermill-demo/internal/infra/messaging/middleware"
	"github.com/rinkachi/golang-demos/golang-watermill-demo/internal/infra/telemetry"
	httpapi "github.com/rinkachi/golang-demos/golang-watermill-demo/internal/interfaces/http"
	"github.com/rinkachi/golang-demos/golang-watermill-demo/internal/interfaces/ws"
)

func main() {
	cfg := config.LoadAPI()
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
	metricsBuilder := metrics.NewPrometheusMetricsBuilder(promRegistry, "watermill", "api")

	kafkaPublisher := messaging.NewKafkaPublisher(cfg.KafkaBrokers, logger)
	kafkaSubscriber := messaging.NewKafkaSubscriber(cfg.KafkaBrokers, "watermill-api", logger)
	rabbitPublisher := messaging.NewRabbitPublisher(cfg.RabbitURL, logger)
	defer kafkaPublisher.Close()
	defer kafkaSubscriber.Close()
	defer rabbitPublisher.Close()

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
		infraMiddleware.Logging(logger),
		infraMiddleware.Tracing(otel.Tracer(cfg.ServiceName), propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		)),
	)

	metricsBuilder.AddPrometheusRouterMetrics(router)
	router.AddPlugin(plugin.SignalsHandler)

	hub := ws.NewHub(logger)
	go hub.Run()

	router.AddConsumerHandler(
		"realtime_ws_fanout",
		topics.RealtimeUpdates,
		kafkaSubscriber,
		func(msg *message.Message) error {
			hub.Broadcast(msg.Payload)
			return nil
		},
	)

	go func() {
		if err := router.Run(ctx); err != nil {
			logger.Error("router_run_failed", err, nil)
		}
	}()

	httpServer := startHTTP(cfg.HTTPAddr, logger, kafkaPublisher, rabbitPublisher, hub)

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = httpServer.Shutdown(shutdownCtx)
	_ = router.Close()
}

func startHTTP(addr string, logger watermill.LoggerAdapter, kafkaPub message.Publisher, rabbitPub message.Publisher, hub *ws.Hub) *http.Server {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())

	handlers := httpapi.NewHandlers(logger, kafkaPub, rabbitPub)
	handlers.Register(router)
	router.GET("/ws", func(c *gin.Context) {
		hub.ServeHTTP(c.Writer, c.Request)
	})

	server := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	go func() {
		logger.Info("http_server_started", watermill.LogFields{
			"addr": addr,
		})
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("http_server_failed", err, nil)
		}
	}()

	return server
}
