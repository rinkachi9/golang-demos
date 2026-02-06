package patterns

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
)

// RateLimiter controls the frequency of events.
type RateLimiter struct {
	tickInterval time.Duration
	ticker       *time.Ticker
	tokens       chan struct{}
}

func NewRateLimiter(rps int, burst int) *RateLimiter {
	interval := time.Second / time.Duration(rps)
	rl := &RateLimiter{
		tickInterval: interval,
		ticker:       time.NewTicker(interval),
		tokens:       make(chan struct{}, burst),
	}

	// Fill burst
	for i := 0; i < burst; i++ {
		rl.tokens <- struct{}{}
	}

	go func() {
		for range rl.ticker.C {
			select {
			case rl.tokens <- struct{}{}:
			default:
				// Bucket full
			}
		}
	}()

	return rl
}

func (rl *RateLimiter) Wait(ctx context.Context) error {
	tracer := otel.Tracer("ratelimiter")
	_, span := tracer.Start(ctx, "ratelimiter_wait")
	defer span.End()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-rl.tokens:
		return nil
	}
}

func (rl *RateLimiter) Stop() {
	rl.ticker.Stop()
}
