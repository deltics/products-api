package api

import (
	"context"
	"products-api/internal/api/ratelimiter"
	"time"
)

// NewRateLimiter initializes a new rate limiter with the specified limit.
// If the limit is less than or equal to zero, it returns a NoopLimiter that
// does not enforce any rate limiting.
func NewRateLimiter(ctx context.Context, limit int) (RateLimiter, error) {
	if limit <= 0 {
		return ratelimiter.NewNoopLimiter(), nil
	}

	// configuration of the rate limiter would be more comprehensive in
	// a real application
	//
	// for this example the per second limit may be specified with a fixed
	// client timeout (1 minute)
	cfg := ratelimiter.Config{
		Limit:         limit,
		LimitInterval: time.Second,
		ClientTimeout: time.Minute,
	}

	// create the rate limiter with the specified configuration
	limiter, err := ratelimiter.New(ctx, cfg)
	if err != nil {
		return nil, err
	}

	return limiter, nil
}
