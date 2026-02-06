module github.com/rinkachi/golang-demos/golang-otel-demo

go 1.22

require (
	github.com/ThreeDotsLabs/watermill v1.3.5
	github.com/ThreeDotsLabs/watermill-kafka/v2 v2.5.0
	github.com/gin-gonic/gin v1.9.1
	github.com/uptrace/opentelemetry-go-extra/otelgorm v0.2.3
	go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin v0.46.1
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.46.1
	go.opentelemetry.io/otel v1.21.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.21.0
	go.opentelemetry.io/otel/exporters/prometheus v0.42.0
	go.opentelemetry.io/otel/sdk v1.21.0
	go.opentelemetry.io/otel/trace v1.21.0
	gorm.io/driver/postgres v1.5.4
	gorm.io/gorm v1.25.5
)
