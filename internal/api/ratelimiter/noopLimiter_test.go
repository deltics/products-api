package ratelimiter_test

import (
	"net/http"
	"products-api/internal/api/ratelimiter"
	"testing"
)

func TestNoopLimiter(t *testing.T) {
	// create a rate limiter with no limit
	rateLimiter := ratelimiter.NewNoopLimiter()

	// all requests should be allowed
	for i := range 1000 {
		if !rateLimiter.Allow(&http.Request{RemoteAddr: "test"}) {
			t.Errorf("Expected request #%d to be allowed", i)
		}
	}
}
