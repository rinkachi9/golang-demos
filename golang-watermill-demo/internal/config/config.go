package config

import (
	"os"
	"strings"
)

type APIConfig struct {
	ServiceName    string
	ServiceVersion string
	HTTPAddr       string
	MetricsAddr    string
	OtelEndpoint   string
	KafkaBrokers   []string
	RabbitURL      string
}

type WorkerConfig struct {
	ServiceName    string
	ServiceVersion string
	MetricsAddr    string
	OtelEndpoint   string
	KafkaBrokers   []string
	RabbitURL      string
}

func LoadAPI() APIConfig {
	return APIConfig{
		ServiceName:    getEnv("SERVICE_NAME", "watermill-api"),
		ServiceVersion: getEnv("SERVICE_VERSION", "0.1.0"),
		HTTPAddr:       getEnv("HTTP_ADDR", ":8085"),
		MetricsAddr:    getEnv("METRICS_ADDR", ":9109"),
		OtelEndpoint:   getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4317"),
		KafkaBrokers:   splitAndTrim(getEnv("KAFKA_BROKERS", "localhost:9092")),
		RabbitURL:      getEnv("RABBIT_URL", "amqp://guest:guest@localhost:5672/"),
	}
}

func LoadWorker() WorkerConfig {
	return WorkerConfig{
		ServiceName:    getEnv("SERVICE_NAME", "watermill-worker"),
		ServiceVersion: getEnv("SERVICE_VERSION", "0.1.0"),
		MetricsAddr:    getEnv("METRICS_ADDR", ":9110"),
		OtelEndpoint:   getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4317"),
		KafkaBrokers:   splitAndTrim(getEnv("KAFKA_BROKERS", "localhost:9092")),
		RabbitURL:      getEnv("RABBIT_URL", "amqp://guest:guest@localhost:5672/"),
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func splitAndTrim(v string) []string {
	raw := strings.Split(v, ",")
	out := make([]string, 0, len(raw))
	for _, item := range raw {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		out = append(out, trimmed)
	}
	return out
}
