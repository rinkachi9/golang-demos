package patterns

import (
	"context"

	"go.opentelemetry.io/otel"
)

// PipelineStage defines a step in the pipeline.
type PipelineStage[T any] func(context.Context, <-chan T) <-chan T

// ... (Generator and Filter are handled in previous block, keeping context valid for Map)

// Map applies a function to all items in the channel.
func Map[T any, R any](ctx context.Context, in <-chan T, transform func(T) R) <-chan R {
	outStr := make(chan R)
	tracer := otel.Tracer("pipeline")

	go func() {
		defer close(outStr)
		_, span := tracer.Start(ctx, "map")
		defer span.End()

		for item := range in {
			mapped := transform(item)
			select {
			case <-ctx.Done():
				return
			case outStr <- mapped:
			}
		}
	}()
	return outStr
}
	outStr := make(chan T)
	tracer := otel.Tracer("pipeline")

	go func() {
		defer close(outStr)
		_, span := tracer.Start(ctx, "generator")
		defer span.End()

		for _, item := range items {
			select {
			case <-ctx.Done():
				return
			case outStr <- item:
			}
		}
	}()
	return outStr
}

// Filter returns a new channel with items that satisfy the predicate.
func Filter[T any](ctx context.Context, in <-chan T, predicate func(T) bool) <-chan T {
	outStr := make(chan T)
	tracer := otel.Tracer("pipeline")

	go func() {
		defer close(outStr)
		_, span := tracer.Start(ctx, "filter")
		defer span.End()

		for item := range in {
			if !predicate(item) {
				continue
			}
			select {
			case <-ctx.Done():
				return
			case outStr <- item:
			}
		}
	}()
	return outStr
}

// Map applies a function to all items in the channel.
func Map[T any, R any](ctx context.Context, in <-chan T, transform func(T) R) <-chan R {
	outStr := make(chan R)
	go func() {
		defer close(outStr)
		for item := range in {
			mapped := transform(item)
			select {
			case <-ctx.Done():
				return
			case outStr <- mapped:
			}
		}
	}()
	return outStr
}
