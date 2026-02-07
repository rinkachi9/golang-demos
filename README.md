# golang-demos

Purpose: a monorepo of small, focused Go demos exploring architecture, concurrency, messaging, observability, resilience, and infrastructure tooling.

## Usage

### Docker Compose

Start shared infrastructure from repo root:
```bash
docker compose up -d
```

Stop and remove containers:
```bash
docker compose down
```

### Dependencies

Most projects use standard Go modules:
```bash
cd <project-dir>
go mod tidy
```

Project-specific system dependencies:
- `golang-ebpf-monitoring`: `clang`, `llvm`, `libbpf-dev` (see `golang-ebpf-monitoring/README.md`)
- `golang-watermill-demo`: Kafka, RabbitMQ, Jaeger (via docker-compose)
- `golang-feature-toggle-poc`: Flipt (via docker-compose)
- `golang-clean-architecture`: PostgreSQL (via docker-compose or local)
- `golang-gorm-advanced`: PostgreSQL (via docker-compose or local)

### Go Work (Workspace)

This repo uses `go.work` to manage the multi-module workspace:
```bash
go work use ./golang-watermill-demo
go work sync
```

Typical workflow:
```bash
go work use ./golang-*
go work sync
go list ./...
```

If you add a new module, append it to `go.work` and run `go work sync`.

## Projects catalog

### `golang-clean-architecture`
* Clean Architecture sample with Gin, Watermill (in‑memory pub/sub), and PostgreSQL persistence.
* Docs: [golang-clean-architecture/README.md](golang-clean-architecture/README.md)

### `golang-concurrency-patterns`
* Concurrency patterns (generator, worker pool, fan‑in, barrier, errgroup) with OpenTelemetry traces.
* Docs: [golang-concurrency-patterns/README.md](golang-concurrency-patterns/README.md)

### `golang-config-hot-reload`
* Hot‑reload config via Viper + fsnotify with a small HTTP endpoint.
* Docs: [golang-config-hot-reload/README.md](golang-config-hot-reload/README.md)

### `golang-ebpf-monitoring`
* eBPF XDP packet counter with `bpf2go`.
* Docs: [golang-ebpf-monitoring/README.md](golang-ebpf-monitoring/README.md)

### `golang-feature-toggle-poc`
* Feature flags using Flipt Go SDK + simple HTTP endpoint.
* Docs: [golang-feature-toggle-poc/README.md](golang-feature-toggle-poc/README.md)

### `golang-generics-demo`
* Go generics helpers: slices, maps, sets, stack, repository, pointers.
* Docs: [golang-generics-demo/README.md](golang-generics-demo/README.md)

### `golang-grpc-kafka-demo`
* gRPC streaming into Kafka (server + consumer).
* Docs: [golang-grpc-kafka-demo/README.md](golang-grpc-kafka-demo/README.md)

### `golang-gorm-advanced`
* Advanced GORM + Postgres demo with Gin API and migrator.
* Docs: [golang-gorm-advanced/README.md](golang-gorm-advanced/README.md)

### `golang-otel-demo`
* OpenTelemetry setup example (traces + metrics).
* Docs: [golang-otel-demo/README.md](golang-otel-demo/README.md)

### `golang-resilience-demo`
* Resilient HTTP client: retries, backoff + circuit breaker.
* Docs: [golang-resilience-demo/README.md](golang-resilience-demo/README.md)

### `golang-watermill-demo`
* Clean‑architecture, two‑service Watermill demo with Kafka, RabbitMQ, WebSocket, OTel, Prometheus.
* Docs: [golang-watermill-demo/README.md](golang-watermill-demo/README.md)
