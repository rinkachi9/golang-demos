package patterns

import (
	"context"
	"sync"

	"go.opentelemetry.io/otel"
)

// Barrier synchronizes a fixed number of parties.
type Barrier struct {
	n      int
	count  int
	mutex  sync.Mutex
	cond   *sync.Cond
}

func NewBarrier(n int) *Barrier {
	b := &Barrier{
		n: n,
	}
	b.cond = sync.NewCond(&b.mutex)
	return b
}

func (b *Barrier) Await(ctx context.Context, id string) {
	tracer := otel.Tracer("barrier")
	_, span := tracer.Start(ctx, "barrier_await")
	defer span.End()

	b.mutex.Lock()
	defer b.mutex.Unlock()

	b.count++
	if b.count == b.n {
		// Last one arrived, wake everyone up
		b.count = 0
		b.cond.Broadcast()
	} else {
		// Wait for others
		b.cond.Wait()
	}
}
