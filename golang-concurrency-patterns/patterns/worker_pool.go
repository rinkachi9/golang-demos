package patterns

import (
	"context"
	"fmt"
	"sync"

	"go.opentelemetry.io/otel"
)

// Task represents a unit of work.
type Task[T any, R any] func(context.Context, T) (R, error)

// Result holds the output of a task or an error.
type Result[R any] struct {
	Value R
	Err   error
}

// WorkerPool implements a generic pool of workers processing inputs concurrently.
// It uses a semaphore pattern to limit concurrency.
func WorkerPool[T any, R any](
	ctx context.Context,
	tasks []T,
	workerFunc Task[T, R],
	concurrency int,
) <-chan Result[R] {
	results := make(chan Result[R], len(tasks))
	var wg sync.WaitGroup

	// Semaphore to limit concurrency
	sem := make(chan struct{}, concurrency)

	// Tracer
	tracer := otel.Tracer("worker-pool")

	go func() {
		defer close(results)
		
		// Span for the pool manager
		ctx, span := tracer.Start(ctx, "worker_pool_manager")
		defer span.End()

		for i, taskInput := range tasks {
			// Check for cancellation early
			select {
			case <-ctx.Done():
				return
			case sem <- struct{}{}: // Acquire token
			}

			wg.Add(1)
			go func(idx int, input T) {
				defer wg.Done()
				defer func() { <-sem }() // Release token

				// Start span for the worker
				_, wSpan := tracer.Start(ctx, fmt.Sprintf("worker_%d", idx))
				defer wSpan.End()

				val, err := workerFunc(ctx, input)
				results <- Result[R]{Value: val, Err: err}
			}(i, taskInput)
		}

		wg.Wait()
	}()

	return results
}
