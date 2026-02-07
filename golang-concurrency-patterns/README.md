# golang-concurrency-patterns

Concurrency patterns demo (generator, worker pool, fan‑in, barrier, errgroup) with OpenTelemetry traces.

## Run

```bash
go run .
```

## Notes

- Expects Jaeger/OTLP on `localhost:4317` (see `docker-compose.yaml` in repo root).
