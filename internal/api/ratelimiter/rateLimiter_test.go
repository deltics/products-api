package ratelimiter_test

import (
	"context"
	"errors"
	"net/http"
	"products-api/internal/api/ratelimiter"
	"testing"

	"github.com/blugnu/time"
)

func TestRateLimiterConfiguration(t *testing.T) {
	ctx := context.Background()
	cfg := ratelimiter.Config{}

	_, err := ratelimiter.New(ctx, cfg)
	if !errors.Is(err, ratelimiter.ErrInvalidLimit) {
		t.Errorf("Expected error for invalid limit, got: %v", err)
	}

	cfg.Limit = 10
	_, err = ratelimiter.New(ctx, cfg)
	if !errors.Is(err, ratelimiter.ErrInvalidLimitInterval) {
		t.Errorf("Expected error for invalid limit interval, got: %v", err)
	}

	cfg.LimitInterval = time.Second
	_, err = ratelimiter.New(ctx, cfg)
	if !errors.Is(err, ratelimiter.ErrInvalidClientTimeout) {
		t.Errorf("Expected error for invalid client timeout, got: %v", err)
	}
}

func TestRateLimiter(t *testing.T) {
	// establish a context with a mock clock for testing
	// this allows us to control time in tests and simulate the passage
	// of time for rate limiting
	clock := time.NewMockClock()
	ctx := context.Background()
	ctx = time.ContextWithClock(ctx, clock)

	// add cancellation to the context to ensure the rate limiter
	// can be stopped after tests
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// setup rate limiter with a limit of 5 requests per second
	// and a client timeout of 1 minute
	cfg := ratelimiter.Config{
		Limit:         5,
		LimitInterval: time.Second,
		ClientTimeout: time.Minute,
	}

	rateLimiter, err := ratelimiter.New(ctx, cfg)
	if err != nil {
		t.Fatalf("Failed to create rate limiter: %v", err)
	}

	// 6 requests will trigger rate limiting
	for i := 1; i <= 6; i++ {
		result := rateLimiter.Allow(&http.Request{
			RemoteAddr: "test",
		})
		switch i {
		case 1, 2, 3, 4, 5:
			if !result {
				t.Errorf("Expected request #%d to be allowed", i)
			}
		case 6:
			if result {
				t.Error("Expected request #5 to be disallowed")
			}
		}
	}

	// simulate the passing of a limit interval
	clock.AdvanceBy(cfg.LimitInterval)

	// a further request should now succeed
	if result := rateLimiter.Allow(&http.Request{RemoteAddr: "test"}); !result {
		t.Error("Expected request to be allowed")
	}

	// simulate the passing of 2 client timeout intervals
	//
	// the client would not timeout in the first interval since it
	// made requests in that time, so the second interval is required,
	// with no client activity, to trigger cleanup
	clock.AdvanceBy(2 * cfg.ClientTimeout)

	// the rate limiter should have cleaned up old clients
	if rateLimiter.NumberOfClients() != 0 {
		t.Errorf("Expected no clients after client timeout, got %d", rateLimiter.NumberOfClients())
	}
}
