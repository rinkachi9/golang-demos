package patterns

import (
	"context"

	"go.opentelemetry.io/otel"
)

// Future represents a value that will be available in the future.
type Future[T any] struct {
	val  T
	err  error
	done chan struct{}
}

// Result blocks until the future is resolved/rejected.
func (f *Future[T]) Result(ctx context.Context) (T, error) {
	select {
	case <-ctx.Done():
		var zero T
		return zero, ctx.Err()
	case <-f.done:
		return f.val, f.err
	}
}

// Async executes a function asynchronously and returns a Future.
func Async[T any](ctx context.Context, f func(context.Context) (T, error)) *Future[T] {
	fut := &Future[T]{
		done: make(chan struct{}),
	}

	tracer := otel.Tracer("future")

	go func() {
		defer close(fut.done)

		// Start span for async task
		subCtx, span := tracer.Start(ctx, "async_future")
		defer span.End()

		fut.val, fut.err = f(subCtx)
		if fut.err != nil {
			span.RecordError(fut.err)
		}
	}()

	return fut
}
