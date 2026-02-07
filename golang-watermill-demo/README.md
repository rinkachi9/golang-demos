# golang-watermill-demo

Clean-architecture demo Watermill z dwiema usługami:

- `api`: Gin HTTP + WebSocket + publikacja do Kafka/Rabbit
- `worker`: konsumuje Kafka/Rabbit, enrichuje dane, publikuje do Kafka i Rabbit

## Szybki start

1. Infrastruktura:
```bash
docker compose up -d kafka zookeeper rabbitmq jaeger
```

2. Start usług:
```bash
cd golang-watermill-demo
go run ./cmd/api
go run ./cmd/worker
```

3. Opublikuj zdarzenie:
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

## Endpointy

- `POST /api/order` - publikuje zdarzenie do Kafki i Rabbit
- `GET /ws` - WebSocket realtime updates
- `GET /healthz` - healthcheck

## Konfiguracja (ENV)

Wspólne:
- `KAFKA_BROKERS` (domyślnie `localhost:9092`)
- `RABBIT_URL` (domyślnie `amqp://guest:guest@localhost:5672/`)
- `OTEL_EXPORTER_OTLP_ENDPOINT` (domyślnie `localhost:4317`)
- `SERVICE_VERSION` (domyślnie `0.1.0`)

API:
- `SERVICE_NAME` (domyślnie `watermill-api`)
- `HTTP_ADDR` (domyślnie `:8085`)
- `METRICS_ADDR` (domyślnie `:9109`)

Worker:
- `SERVICE_NAME` (domyślnie `watermill-worker`)
- `METRICS_ADDR` (domyślnie `:9110`)

## RabbitMQ Dead-Letter

Worker ustawia DLX:
- exchange: `watermill.dlx`
- queue: `watermill.dead_letter`
- routing key: `dead-letter`

Wiadomości z handlerów po wyczerpaniu retry trafiają do DLQ.
