package sync

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"
)

// BackoffConfig holds configuration for exponential backoff with jitter.
type BackoffConfig struct {
	InitialInterval time.Duration
	Multiplier      float64
	MaxInterval     time.Duration
	JitterFraction  float64 // 0.0 to 1.0
}

// DefaultBackoffConfig returns a sensible default backoff configuration.
func DefaultBackoffConfig() BackoffConfig {
	return BackoffConfig{
		InitialInterval: 250 * time.Millisecond,
		Multiplier:      2.0,
		MaxInterval:     30 * time.Second,
		JitterFraction:  0.2,
	}
}

// Backoff computes the next wait duration for a given attempt (0-indexed)
// using exponential backoff capped at MaxInterval, with optional jitter.
func (c BackoffConfig) Backoff(attempt int) time.Duration {
	if c.Multiplier <= 1.0 {
		c.Multiplier = 2.0
	}
	base := float64(c.InitialInterval) * math.Pow(c.Multiplier, float64(attempt))
	if base > float64(c.MaxInterval) {
		base = float64(c.MaxInterval)
	}
	if c.JitterFraction > 0 {
		jitter := base * c.JitterFraction * (rand.Float64()*2 - 1)
		base += jitter
		if base < 0 {
			base = 0
		}
	}
	return time.Duration(base)
}

// Validate checks that the BackoffConfig has sensible values.
func (c BackoffConfig) Validate() error {
	if c.InitialInterval <= 0 {
		return fmt.Errorf("backoff: InitialInterval must be positive, got %v", c.InitialInterval)
	}
	if c.MaxInterval < c.InitialInterval {
		return fmt.Errorf("backoff: MaxInterval (%v) must be >= InitialInterval (%v)", c.MaxInterval, c.InitialInterval)
	}
	if c.JitterFraction < 0 || c.JitterFraction > 1 {
		return fmt.Errorf("backoff: JitterFraction must be in [0, 1], got %v", c.JitterFraction)
	}
	return nil
}

// Sleep waits for the backoff duration for the given attempt, or until ctx
// is cancelled. Returns ctx.Err() if the context was cancelled.
func (c BackoffConfig) Sleep(ctx context.Context, attempt int) error {
	d := c.Backoff(attempt)
	if d <= 0 {
		return nil
	}
	select {
	case <-time.After(d):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
