package ratelimiter

import "net/http"

type NoopLimiter struct{}

func NewNoopLimiter() *NoopLimiter {
	return &NoopLimiter{}
}

// Allow always returns true, indicating that all requests are allowed
func (n *NoopLimiter) Allow(rq *http.Request) bool {
	return true
}
