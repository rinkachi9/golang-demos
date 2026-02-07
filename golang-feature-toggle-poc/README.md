# golang-feature-toggle-poc

Feature flag demo using Flipt Go SDK with a simple HTTP endpoint.

## Run

1. Start Flipt (see root `docker-compose.yaml`).
2. Run:
```bash
go run .
```

## Endpoint

- `GET /feature` - evaluates `new-ui-enabled` flag.
