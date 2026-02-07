# golang-grpc-kafka-demo

gRPC streaming service that publishes to Kafka, plus a consumer.

## Run

1. Start Kafka (see root `docker-compose.yaml`).
2. Generate gRPC code from `proto/river.proto` into `pkg/api/v1` (the repo does not include generated code).
3. Run server:
```bash
go run ./cmd/server
```
4. Run consumer:
```bash
go run ./cmd/consumer
```

## Notes

- The server listens on `:50051` and publishes to topic `sensor-data`.
