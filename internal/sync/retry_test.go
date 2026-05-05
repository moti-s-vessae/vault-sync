package sync_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/your-org/vault-sync/internal/sync"
	"github.com/your-org/vault-sync/internal/vault"
)

type callCountLoader struct {
	calls  int
	errs   []error
	result map[string]string
}

func (c *callCountLoader) Load(_ context.Context, _ string) (map[string]string, error) {
	idx := c.calls
	c.calls++
	if idx < len(c.errs) {
		return nil, c.errs[idx]
	}
	return c.result, nil
}

func TestNewRetryLoader_InvalidArgs(t *testing.T) {
	inner := &callCountLoader{}
	if _, err := sync.NewRetryLoader(nil, 1, time.Millisecond); err == nil {
		t.Error("expected error for nil inner")
	}
	if _, err := sync.NewRetryLoader(inner, 0, time.Millisecond); err == nil {
		t.Error("expected error for maxRetry=0")
	}
	if _, err := sync.NewRetryLoader(inner, 1, 0); err == nil {
		t.Error("expected error for baseWait=0")
	}
}

func TestRetryLoader_SuccessOnFirstAttempt(t *testing.T) {
	inner := &callCountLoader{result: map[string]string{"K": "V"}}
	rl, err := sync.NewRetryLoader(inner, 3, time.Millisecond)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, err := rl.Load(context.Background(), "secret/data/app")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got["K"] != "V" {
		t.Errorf("expected V, got %s", got["K"])
	}
	if inner.calls != 1 {
		t.Errorf("expected 1 call, got %d", inner.calls)
	}
}

func TestRetryLoader_RetriesTransientError(t *testing.T) {
	transient := errors.New("connection reset")
	inner := &callCountLoader{
		errs:   []error{transient, transient},
		result: map[string]string{"X": "1"},
	}
	rl, _ := sync.NewRetryLoader(inner, 3, time.Millisecond)
	got, err := rl.Load(context.Background(), "secret/data/app")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got["X"] != "1" {
		t.Errorf("expected 1, got %s", got["X"])
	}
	if inner.calls != 3 {
		t.Errorf("expected 3 calls, got %d", inner.calls)
	}
}

func TestRetryLoader_ExhaustsRetries(t *testing.T) {
	transient := errors.New("timeout")
	inner := &callCountLoader{errs: []error{transient, transient, transient, transient}}
	rl, _ := sync.NewRetryLoader(inner, 2, time.Millisecond)
	_, err := rl.Load(context.Background(), "secret/data/app")
	if err == nil {
		t.Fatal("expected error after exhausting retries")
	}
	if inner.calls != 3 { // 1 initial + 2 retries
		t.Errorf("expected 3 calls, got %d", inner.calls)
	}
}

func TestRetryLoader_NoRetryOnNotFound(t *testing.T) {
	inner := &callCountLoader{errs: []error{vault.ErrNotFound}}
	rl, _ := sync.NewRetryLoader(inner, 3, time.Millisecond)
	_, err := rl.Load(context.Background(), "secret/data/missing")
	if !errors.Is(err, vault.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
	if inner.calls != 1 {
		t.Errorf("expected 1 call (no retry), got %d", inner.calls)
	}
}

func TestRetryLoader_ContextCancelled(t *testing.T) {
	transient := errors.New("transient")
	inner := &callCountLoader{errs: []error{transient, transient, transient}}
	rl, _ := sync.NewRetryLoader(inner, 5, 50*time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately
	_, err := rl.Load(ctx, "secret/data/app")
	if err == nil {
		t.Fatal("expected error on cancelled context")
	}
}
