# golang-otel-demo

OpenTelemetry setup example for traces and Prometheus metrics.

## Notes

- `telemetry.go` provides a reusable `setupOTelSDK` helper.
- Expects an OTLP endpoint on `localhost:4317`.
- Prometheus config is in `prometheus.yml` (scrape target uses `host.docker.internal`).
