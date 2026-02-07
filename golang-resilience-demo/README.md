# golang-resilience-demo

Resilient HTTP client demo: retries, exponential backoff with jitter, and a circuit breaker.

## Run

```bash
go run .
```

## Notes

- Starts a flaky server on `:8082` and issues client requests against it.
