package sync

import (
	"context"
	"fmt"
	"time"
)

// SecretsLoader is the interface for loading secrets from a path.
type SecretsLoader interface {
	Load(ctx context.Context, path string) (map[string]string, error)
}

// TimeoutLoader wraps a SecretsLoader and enforces a per-call deadline.
type TimeoutLoader struct {
	inner   SecretsLoader
	timeout time.Duration
}

// NewTimeoutLoader creates a TimeoutLoader that cancels any Load call
// that exceeds the given timeout. Returns an error if inner is nil or
// timeout is non-positive.
func NewTimeoutLoader(inner SecretsLoader, timeout time.Duration) (*TimeoutLoader, error) {
	if inner == nil {
		return nil, fmt.Errorf("timeout_loader: inner loader must not be nil")
	}
	if timeout <= 0 {
		return nil, fmt.Errorf("timeout_loader: timeout must be positive, got %v", timeout)
	}
	return &TimeoutLoader{inner: inner, timeout: timeout}, nil
}

// Load calls the inner loader with a derived context that times out after
// the configured duration. If the parent context is already cancelled the
// error propagates immediately.
func (t *TimeoutLoader) Load(ctx context.Context, path string) (map[string]string, error) {
	ctx, cancel := context.WithTimeout(ctx, t.timeout)
	defer cancel()

	results, err := t.inner.Load(ctx, path)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("timeout_loader: load %q timed out after %v", path, t.timeout)
		}
		return nil, err
	}
	return results, nil
}
