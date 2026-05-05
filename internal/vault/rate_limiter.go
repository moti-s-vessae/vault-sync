package vault

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// RateLimiter enforces a maximum number of requests per second to Vault.
type RateLimiter struct {
	mu       sync.Mutex
	tokens   float64
	max      float64
	refill   float64 // tokens per second
	lastTime time.Time
}

// NewRateLimiter creates a RateLimiter allowing rps requests per second.
func NewRateLimiter(rps float64) (*RateLimiter, error) {
	if rps <= 0 {
		return nil, fmt.Errorf("rps must be positive, got %f", rps)
	}
	return &RateLimiter{
		tokens:   rps,
		max:      rps,
		refill:   rps,
		lastTime: time.Now(),
	}, nil
}

// Wait blocks until a token is available or ctx is cancelled.
func (r *RateLimiter) Wait(ctx context.Context) error {
	for {
		r.mu.Lock()
		now := time.Now()
		elapsed := now.Sub(r.lastTime).Seconds()
		r.lastTime = now
		r.tokens += elapsed * r.refill
		if r.tokens > r.max {
			r.tokens = r.max
		}
		if r.tokens >= 1.0 {
			r.tokens -= 1.0
			r.mu.Unlock()
			return nil
		}
		wait := time.Duration((1.0-r.tokens)/r.refill*float64(time.Second))
		r.mu.Unlock()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(wait):
		}
	}
}

// Available returns the current token count (approximate).
func (r *RateLimiter) Available() float64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.tokens
}
