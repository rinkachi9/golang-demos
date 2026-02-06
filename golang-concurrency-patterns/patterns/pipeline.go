package patterns

import (
	"context"

	"go.opentelemetry.io/otel"
)

// PipelineStage defines a step in the pipeline.
type PipelineStage[T any] func(context.Context, <-chan T) <-chan T

// Generator emits provided items into a channel.
func Generator[T any](ctx context.Context, items ...T) <-chan T {
	outCh := make(chan T)
	tracer := otel.Tracer("pipeline")

	go func() {
		defer close(outCh)
		_, span := tracer.Start(ctx, "generator")
		defer span.End()

		for _, item := range items {
			select {
			case <-ctx.Done():
				return
			case outCh <- item:
			}
		}
	}()

	return outCh
}

// Filter returns a new channel with items that satisfy the predicate.
func Filter[T any](ctx context.Context, in <-chan T, predicate func(T) bool) <-chan T {
	outCh := make(chan T)
	tracer := otel.Tracer("pipeline")

	go func() {
		defer close(outCh)
		_, span := tracer.Start(ctx, "filter")
		defer span.End()

		for item := range in {
			if !predicate(item) {
				continue
			}
			select {
			case <-ctx.Done():
				return
			case outCh <- item:
			}
		}
	}()

	return outCh
}

// Map applies a function to all items in the channel.
func Map[T any, R any](ctx context.Context, in <-chan T, transform func(T) R) <-chan R {
	outCh := make(chan R)
	tracer := otel.Tracer("pipeline")

	go func() {
		defer close(outCh)
		_, span := tracer.Start(ctx, "map")
		defer span.End()

		for item := range in {
			mapped := transform(item)
			select {
			case <-ctx.Done():
				return
			case outCh <- mapped:
			}
		}
	}()

	return outCh
}
