package httpclient

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"

	"github.com/sony/gobreaker"
)

// ResilientClient wraps http.Client with retry and circuit breaker logic.
type ResilientClient struct {
	client *http.Client
	cb     *gobreaker.CircuitBreaker
}

func NewResilientClient() *ResilientClient {
	// Circuit Breaker Configuration
	st := gobreaker.Settings{
		Name:        "HTTPClient",
		MaxRequests: 3,
		Interval:    5 * time.Second,
		Timeout:     10 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 3 && failureRatio >= 0.6
		},
	}

	return &ResilientClient{
		client: &http.Client{Timeout: 5 * time.Second},
		cb:     gobreaker.NewCircuitBreaker(st),
	}
}

// Get performs a GET request with retries and circuit breaker protection.
func (c *ResilientClient) Get(ctx context.Context, url string) ([]byte, error) {
	var body []byte

	// Executing request inside Circuit Breaker
	_, err := c.cb.Execute(func() (interface{}, error) {
		return c.getWithRetry(ctx, url)
	})

	if err != nil {
		return nil, err
	}

	return body, nil
}

// getWithRetry implements Exponential Backoff with Jitter
func (c *ResilientClient) getWithRetry(ctx context.Context, url string) ([]byte, error) {
	const maxRetries = 3
	baseDelay := 100 * time.Millisecond

	for i := 0; i <= maxRetries; i++ {
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return nil, err
		}

		resp, err := c.client.Do(req)
		
		// If request succeeds
		if err == nil {
			if resp.StatusCode < 500 {
				defer resp.Body.Close()
				return io.ReadAll(resp.Body)
			}
			// Treat 5xx as retryable failure
			resp.Body.Close()
			err = fmt.Errorf("server error: %d", resp.StatusCode)
		}

		// If this was the last attempt, return error
		if i == maxRetries {
			return nil, err
		}

		// Calculate backoff: base * 2^i + jitter
		backoff := baseDelay * time.Duration(1<<i)
		jitter := (rand.Float64()*0.5 + 0.5) * float64(backoff) // 0.5-1.0 * backoff
		sleepDuration := time.Duration(jitter)

		fmt.Printf("Request failed (%v). Retrying in %v...\n", err, sleepDuration)
		
		select {
		case <-time.After(sleepDuration):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	return nil, errors.New("unreachable")
}
