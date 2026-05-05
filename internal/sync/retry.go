package sync

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/your-org/vault-sync/internal/vault"
)

// RetryLoader wraps a SecretsLoader and retries transient failures with
// exponential backoff.
type RetryLoader struct {
	inner    SecretsLoader
	maxRetry int
	baseWait time.Duration
}

// SecretsLoader is the interface satisfied by any loader that fetches secrets.
type SecretsLoader interface {
	Load(ctx context.Context, path string) (map[string]string, error)
}

// NewRetryLoader returns a RetryLoader that retries up to maxRetry times,
// starting with baseWait between attempts (doubled each retry).
// maxRetry must be >= 1 and baseWait must be > 0.
func NewRetryLoader(inner SecretsLoader, maxRetry int, baseWait time.Duration) (*RetryLoader, error) {
	if inner == nil {
		return nil, errors.New("retry: inner loader must not be nil")
	}
	if maxRetry < 1 {
		return nil, fmt.Errorf("retry: maxRetry must be >= 1, got %d", maxRetry)
	}
	if baseWait <= 0 {
		return nil, fmt.Errorf("retry: baseWait must be > 0, got %s", baseWait)
	}
	return &RetryLoader{inner: inner, maxRetry: maxRetry, baseWait: baseWait}, nil
}

// Load calls the inner loader, retrying on transient vault errors up to
// maxRetry times with exponential backoff. Non-retryable errors (e.g.
// ErrNotFound) are returned immediately.
func (r *RetryLoader) Load(ctx context.Context, path string) (map[string]string, error) {
	wait := r.baseWait
	var lastErr error
	for attempt := 0; attempt <= r.maxRetry; attempt++ {
		secrets, err := r.inner.Load(ctx, path)
		if err == nil {
			return secrets, nil
		}
		// Do not retry on non-transient errors.
		if errors.Is(err, vault.ErrNotFound) || errors.Is(err, vault.ErrForbidden) {
			return nil, err
		}
		lastErr = err
		if attempt < r.maxRetry {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(wait):
			}
			wait *= 2
		}
	}
	return nil, fmt.Errorf("retry: all %d attempts failed: %w", r.maxRetry+1, lastErr)
}
