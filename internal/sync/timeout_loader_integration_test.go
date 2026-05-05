package sync

import (
	"context"
	"strings"
	"testing"
	"time"
)

// TestTimeoutLoader_Integration_WrapsRetryLoader verifies that TimeoutLoader
// can wrap a RetryLoader and that a slow inner loader is cancelled correctly
// even when the retry loop is running.
func TestTimeoutLoader_Integration_WrapsRetryLoader(t *testing.T) {
	// slowLoader that always times out — retry will keep trying until the
	// TimeoutLoader's deadline fires.
	inner := &slowLoader{delay: 500 * time.Millisecond}

	retryL, err := NewRetryLoader(inner, 5, 5*time.Millisecond)
	if err != nil {
		t.Fatalf("NewRetryLoader: %v", err)
	}

	timeoutL, err := NewTimeoutLoader(retryL, 80*time.Millisecond)
	if err != nil {
		t.Fatalf("NewTimeoutLoader: %v", err)
	}

	start := time.Now()
	_, loadErr := timeoutL.Load(context.Background(), "secret/app")
	elapsed := time.Since(start)

	if loadErr == nil {
		t.Fatal("expected an error from timeout, got nil")
	}
	if elapsed > 300*time.Millisecond {
		t.Errorf("load took too long (%v), timeout not enforced", elapsed)
	}
	if !strings.Contains(loadErr.Error(), "timed out") && !strings.Contains(loadErr.Error(), "context") {
		t.Errorf("unexpected error message: %v", loadErr)
	}
}
