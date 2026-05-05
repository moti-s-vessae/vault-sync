package sync

import (
	"context"
	"fmt"

	"github.com/your-org/vault-sync/internal/vault"
)

// SecretLoader is the interface for loading secrets from a path.
type SecretLoader interface {
	Load(ctx context.Context, path string) (map[string]string, error)
}

// RateLimitedLoader wraps a SecretLoader and enforces a rate limit.
type RateLimitedLoader struct {
	inner   SecretLoader
	limiter *vault.RateLimiter
}

// NewRateLimitedLoader creates a RateLimitedLoader with the given rps limit.
func NewRateLimitedLoader(inner SecretLoader, rps float64) (*RateLimitedLoader, error) {
	if inner == nil {
		return nil, fmt.Errorf("inner loader must not be nil")
	}
	limiter, err := vault.NewRateLimiter(rps)
	if err != nil {
		return nil, fmt.Errorf("creating rate limiter: %w", err)
	}
	return &RateLimitedLoader{inner: inner, limiter: limiter}, nil
}

// Load waits for a rate-limit token then delegates to the inner loader.
func (r *RateLimitedLoader) Load(ctx context.Context, path string) (map[string]string, error) {
	if err := r.limiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limiter: %w", err)
	}
	return r.inner.Load(ctx, path)
}
