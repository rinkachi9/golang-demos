package patterns

import (
	"context"
	"fmt"
	"sync"
	
	"go.opentelemetry.io/otel"
)

// FanIn merges multiple channels into a single channel.
func FanIn[T any](ctx context.Context, channels ...<-chan T) <-chan T {
	var wg sync.WaitGroup
	out := make(chan T)
	tracer := otel.Tracer("fan-in")
	
	ctx, span := tracer.Start(ctx, "fan_in_manager")
	defer span.End()

	// Multiplex
	multiplex := func(idx int, c <-chan T) {
		defer wg.Done()
		
		_, mSpan := tracer.Start(ctx, fmt.Sprintf("fan_in_channel_%d", idx))
		defer mSpan.End()

		for i := range c {
			select {
			case <-ctx.Done():
				return
			case out <- i:
			}
		}
	}

	wg.Add(len(channels))
	for i, c := range channels {
		go multiplex(i, c)
	}

	// Wait for all to finish in a separate goroutine
	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}
