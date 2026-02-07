# golang-watermill-demo

Clean‑architecture Watermill demo with two services:

- `api`: Gin HTTP + WebSocket + publishes to Kafka and RabbitMQ
- `worker`: consumes Kafka/Rabbit, enriches data, publishes to Kafka and Rabbit

## Quick Start

1. Infrastructure:
```bash
docker compose up -d kafka zookeeper rabbitmq jaeger
```

2. Run services:
```bash
cd golang-watermill-demo
go run ./cmd/api
go run ./cmd/worker
```

3. Publish an event:
```bash
curl -X POST http://localhost:8085/api/order \
  -H 'Content-Type: application/json' \
  -H 'X-Correlation-ID: demo-123' \
  -d '{"id":"order-1","customer":"Ava","total":199.90}'
```

4. WebSocket realtime:
```bash
websocat ws://localhost:8085/ws
```

## Endpoints

- `POST /api/order` - publishes to Kafka and RabbitMQ
- `GET /ws` - WebSocket realtime updates
- `GET /healthz` - healthcheck

## Configuration (ENV)

Shared:
- `KAFKA_BROKERS` (default `localhost:9092`)
- `RABBIT_URL` (default `amqp://guest:guest@localhost:5672/`)
- `OTEL_EXPORTER_OTLP_ENDPOINT` (default `localhost:4317`)
- `SERVICE_VERSION` (default `0.1.0`)

API:
- `SERVICE_NAME` (default `watermill-api`)
- `HTTP_ADDR` (default `:8085`)
- `METRICS_ADDR` (default `:9109`)

Worker:
- `SERVICE_NAME` (default `watermill-worker`)
- `METRICS_ADDR` (default `:9110`)

## RabbitMQ Dead‑Letter

Worker configures DLX:
- exchange: `watermill.dlx`
- queue: `watermill.dead_letter`
- routing key: `dead-letter`

Messages that exhaust retries are routed to the DLQ.
