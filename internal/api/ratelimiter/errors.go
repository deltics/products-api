package ratelimiter

import "errors"

var (
	ErrInvalidLimit         = errors.New("rate limit must be greater than zero")
	ErrInvalidLimitInterval = errors.New("limit interval must be at least one second")
	ErrInvalidClientTimeout = errors.New("client timeout must be greater than limit interval")
)
