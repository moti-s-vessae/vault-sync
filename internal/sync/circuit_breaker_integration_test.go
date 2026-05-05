package sync

import (
	"context"
	"errors"
	"testing"
	"time"
)

// TestCircuitBreaker_IntegrationWithRetryLoader verifies that a CircuitBreaker
// wrapping a RetryLoader opens correctly when the underlying vault is down,
// then recovers once it comes back up.
func TestCircuitBreaker_IntegrationWithRetryLoader(t *testing.T) {
	const (
		threshold = 2
		cooldown  = 20 * time.Millisecond
	)

	flaky := &stubLoader{err: errors.New("vault down")}

	retryLoader, err := NewRetryLoader(flaky, 1, 0)
	if err != nil {
		t.Fatalf("NewRetryLoader: %v", err)
	}

	cb, err := NewCircuitBreaker(retryLoader, threshold, cooldown)
	if err != nil {
		t.Fatalf("NewCircuitBreaker: %v", err)
	}

	// Exhaust threshold — each Load goes through retry (1 retry) then fails.
	for i := 0; i < threshold; i++ {
		_, loadErr := cb.Load(context.Background(), "secret/app")
		if loadErr == nil {
			t.Fatalf("call %d: expected error", i)
		}
	}

	if cb.State() != "open" {
		t.Fatalf("expected open state after %d failures, got %s", threshold, cb.State())
	}

	// Calls while open must return ErrCircuitOpen immediately.
	_, openErr := cb.Load(context.Background(), "secret/app")
	if !errors.Is(openErr, ErrCircuitOpen) {
		t.Fatalf("expected ErrCircuitOpen, got %v", openErr)
	}

	// Wait for cooldown, then fix the inner loader.
	time.Sleep(cooldown + 5*time.Millisecond)
	flaky.err = nil
	flaky.result = map[string]string{"RECOVERED": "true"}

	got, err := cb.Load(context.Background(), "secret/app")
	if err != nil {
		t.Fatalf("expected success after recovery, got %v", err)
	}
	if got["RECOVERED"] != "true" {
		t.Errorf("unexpected result: %v", got)
	}
	if cb.State() != "closed" {
		t.Errorf("expected closed after recovery, got %s", cb.State())
	}
}
