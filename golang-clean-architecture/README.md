# golang-clean-architecture

Clean Architecture demo using Gin, Watermill (in‑memory pub/sub), and PostgreSQL persistence.

## Run

1. Ensure PostgreSQL is running and create database `clean_arch`.
2. Start the app:
```bash
go run ./cmd/app
```

## Notes

- The app connects to Postgres at `localhost:5432` with `user/password`.
- HTTP server listens on `:8080`.
