package patterns

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
	"testing"
	"time"
)

func TestWorkerPool(t *testing.T) {
	ctx := context.Background()
	inputs := []int{1, 2, 3, 4, 5}
	
	// Task: Square the number
	worker := func(ctx context.Context, n int) (int, error) {
		time.Sleep(10 * time.Millisecond) // Simulate work
		return n * n, nil
	}

	resultsChan := WorkerPool(ctx, inputs, worker, 2)
	
	var outputs []int
	for res := range resultsChan {
		if res.Err != nil {
			t.Errorf("unexpected error: %v", res.Err)
		}
		outputs = append(outputs, res.Value)
	}

	sort.Ints(outputs)
	expected := []int{1, 4, 9, 16, 25}
	
	if len(outputs) != len(expected) {
		t.Fatalf("expected length %d, got %d", len(expected), len(outputs))
	}
	
	for i, v := range outputs {
		if v != expected[i] {
			t.Errorf("at index %d: expected %d, got %d", i, expected[i], v)
		}
	}
}

func TestWorkerPool_ErrorHandling(t *testing.T) {
	ctx := context.Background()
	inputs := []int{1, 2}
	
	worker := func(ctx context.Context, n int) (int, error) {
		if n == 2 {
			return 0, errors.New("fail")
		}
		return n, nil
	}

	resultsChan := WorkerPool(ctx, inputs, worker, 2)
	
	var errorsCount int
	for res := range resultsChan {
		if res.Err != nil {
			errorsCount++
		}
	}

	if errorsCount != 1 {
		t.Errorf("expected 1 error, got %d", errorsCount)
	}
}

func TestPipeline(t *testing.T) {
	ctx := context.Background()
	
	// Generator
	input := Generator(ctx, 1, 2, 3, 4, 5, 6)
	
	// Filter: Keep events
	evens := Filter(ctx, input, func(n int) bool {
		return n%2 == 0
	})
	
	// Map: Double
	doubled := Map(ctx, evens, func(n int) int {
		return n * 2
	})
	
	var res []int
	for n := range doubled {
		res = append(res, n)
	}
	
	// 2, 4, 6 -> 4, 8, 12
	expected := []int{4, 8, 12}
	sort.Ints(res)
	
	if len(res) != len(expected) {
		t.Fatalf("expected len %d, got %d", len(expected), len(res))
	}
	for i, v := range res {
		if v != expected[i] {
			t.Errorf("expected %d, got %d", expected[i], v)
		}
	}
}

func TestFanIn(t *testing.T) {
	ctx := context.Background()
	
	c1 := Generator(ctx, 1, 3)
	c2 := Generator(ctx, 2, 4)
	
	merged := FanIn(ctx, c1, c2)
	
	var res []int
	for n := range merged {
		res = append(res, n)
	}
	sort.Ints(res)
	
	expected := []int{1, 2, 3, 4}
	for i, v := range res {
		if v != expected[i] {
			t.Errorf("expected %d, got %d", expected[i], v)
		}
	}
}

func TestCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	
	// Infinite generator
	gen := make(chan int)
	go func() {
		defer close(gen)
		i := 0
		for {
			select {
			case <-ctx.Done():
				return
			case gen <- i:
				i++
			}
		}
	}()
	
	// Read a few items then cancel
	out := Map(ctx, gen, func(n int) int { return n })
	
	count := 0
	for range out {
		count++
		if count == 5 {
			cancel()
			// Should exit loop shortly after cancel
		}
		if count > 100 {
			t.Fatal("Pipeline did not cancel")
		}
	}
}
