package ratelimiter

import (
	"context"
	"net/http"
	"regexp"
	"sync"

	"github.com/blugnu/time"
)

var (
	// a regex to extract the client IP from the request; assumes that
	// the request.RemoteAddr is in the format "IP:port"
	//
	// handles both IPv4 and IPv6 addresses (naively)
	patIP = regexp.MustCompile(`^(.*):[0-9]{1,5}$`)
)

// ClientActivity tracks the number of requests and the last seen time for each client
type ClientActivity struct {
	requestCount int
	lastSeen     time.Time
}

// Config provides configuration for a RateLimiter
type Config struct {
	Limit         int           // Maximum requests per second
	LimitInterval time.Duration // Time interval for the limit
	ClientTimeout time.Duration // Time after which a client is considered inactive
}

// RateLimiter implements a simple rate limiting mechanism
// It tracks the number of requests from each client and allows or denies requests
// based on a configured limit and interval.
type RateLimiter struct {
	sync.RWMutex
	time     time.Clock
	limit    int
	activity map[string]ClientActivity
}

// New creates a new RateLimiter with the specified configuration.
// It validates the configuration and initializes the rate limiter.
// Returns an error if the configuration is invalid.
func New(ctx context.Context, cfg Config) (*RateLimiter, error) {
	if cfg.Limit <= 0 {
		return nil, ErrInvalidLimit
	}
	if cfg.LimitInterval < time.Second {
		return nil, ErrInvalidLimitInterval
	}
	if cfg.ClientTimeout <= cfg.LimitInterval {
		return nil, ErrInvalidClientTimeout
	}

	limiter := &RateLimiter{
		time:     time.ClockFromContext(ctx),
		limit:    cfg.Limit,
		activity: map[string]ClientActivity{},
	}

	limiter.startLimitReset(ctx, cfg.LimitInterval)
	limiter.startClientCleanup(ctx, cfg.ClientTimeout)

	return limiter, nil
}

// Allow returns true if the specified request is allowed to execute.
// It checks if the request from the client is within the allowed
// rate limit.
func (rl *RateLimiter) Allow(rq *http.Request) bool {
	rl.Lock()
	defer rl.Unlock()

	var id = ""
	if patIP.MatchString(rq.RemoteAddr) {
		id = patIP.FindStringSubmatch(rq.RemoteAddr)[1]
	}

	activity, exists := rl.activity[id]
	if !exists {
		activity = ClientActivity{requestCount: 0}
	}

	activity.requestCount += 1
	activity.lastSeen = rl.time.Now()

	rl.activity[id] = activity

	return activity.requestCount <= rl.limit
}

// NumberOfClients returns the number of clients currently tracked by the rate limiter.
// This is useful for monitoring and debugging purposes.
func (rl *RateLimiter) NumberOfClients() int {
	rl.RLock()
	defer rl.RUnlock()

	return len(rl.activity)
}

// startLimitReset starts a goroutine that resets the request count for all clients
// when the configured limit interval expires.
func (rl *RateLimiter) startLimitReset(ctx context.Context, dur time.Duration) {
	ticker := rl.time.NewTicker(dur)
	go func() {
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return

			case <-ticker.C:
				rl.Lock()
				for client, activity := range rl.activity {
					// reset request count for each client
					activity.requestCount = 0
					rl.activity[client] = activity
				}
				rl.Unlock()
			}
		}
	}()
}

// startClientCleanup starts a goroutine that removes clients that have not made
// any requests in the configured client timeout interval.
func (rl *RateLimiter) startClientCleanup(ctx context.Context, dur time.Duration) {
	ticker := rl.time.NewTicker(dur)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return

			case now := <-ticker.C:
				rl.Lock()
				for client, activity := range rl.activity {
					if now.Sub(activity.lastSeen) >= dur {
						delete(rl.activity, client) // remove client if no requests in last 10 seconds
					}
				}
				rl.Unlock()
			}
		}
	}()
}
