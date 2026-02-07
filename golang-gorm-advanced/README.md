# golang-gorm-advanced

Advanced GORM + PostgreSQL demo with a Gin API and a separate migrator.

## Components

- `cmd/api`: Gin REST API with CRUD, transactions, preloading, and query scopes.
- `cmd/migrator`: migration runner with `up`, `down`, `reset`, and `status`.

## Quick Start

1. Start Postgres (root `docker-compose.yaml`):
```bash
docker compose up -d postgres
```

2. Run migrations:
```bash
go run ./cmd/migrator up
```

3. Run API:
```bash
go run ./cmd/api
```

## API Endpoints

- `POST /users`
- `GET /users`
- `GET /users/:id`
- `PATCH /users/:id/deactivate`
- `POST /orders`
- `GET /orders`
- `GET /orders/:id`
- `GET /healthz`
- `GET /metrics` (on `METRICS_ADDR`, default `:9111`)

## Example Requests

Create user:
```bash
curl -X POST http://localhost:8086/users \
  -H 'Content-Type: application/json' \
  -d '{"email":"alice@example.com","name":"Alice"}'
```

Create order:
```bash
curl -X POST http://localhost:8086/orders \
  -H 'Content-Type: application/json' \
  -d '{"user_id":1,"status":"new","items":[{"sku":"sku-1","qty":2,"price":19.99}]}'
```

List users with scopes:
```bash
curl "http://localhost:8086/users?active=true&domain=example.com"
```

List orders with scopes:
```bash
curl "http://localhost:8086/orders?min_total=10&recent_days=7"
```

## Configuration (ENV)

- `DATABASE_URL` (default `host=localhost user=user password=password dbname=gorm_advanced port=5432 sslmode=disable`)
- `HTTP_ADDR` (default `:8086`)
- `METRICS_ADDR` (default `:9111`)

## Migrations

```bash
go run ./cmd/migrator status
go run ./cmd/migrator up
go run ./cmd/migrator down
go run ./cmd/migrator reset
```
