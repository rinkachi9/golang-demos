package patterns

import (
	"context"
	"sync"

	"go.opentelemetry.io/otel"
)

// Group is a collection of goroutines working on subtasks that are part of the same overall task.
type Group struct {
	cancel func()
	wg     sync.WaitGroup
	errOnce sync.Once
	err     error
}

func WithContext(ctx context.Context) (*Group, context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	return &Group{cancel: cancel}, ctx
}

func (g *Group) Go(ctx context.Context, f func() error) {
	g.wg.Add(1)
	tracer := otel.Tracer("errgroup")

	go func() {
		defer g.wg.Done()
		
		// Create a span for this goroutine
		_, span := tracer.Start(ctx, "errgroup_task")
		defer span.End()

		if err := f(); err != nil {
			g.errOnce.Do(func() {
				g.err = err
				if g.cancel != nil {
					g.cancel()
				}
			})
			span.RecordError(err)
		}
	}()
}

func (g *Group) Wait() error {
	g.wg.Wait()
	if g.cancel != nil {
		g.cancel()
	}
	return g.err
}
