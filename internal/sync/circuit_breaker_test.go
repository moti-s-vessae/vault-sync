package sync

import (
	"context"
	"errors"
	"testing"
	"time"
)

type stubLoader struct {
	calls  int
	err    error
	result map[string]string
}

func (s *stubLoader) Load(_ context.Context, _ string) (map[string]string, error) {
	s.calls++
	if s.err != nil {
		return nil, s.err
	}
	return s.result, nil
}

func TestNewCircuitBreaker_InvalidArgs(t *testing.T) {
	inner := &stubLoader{}
	if _, err := NewCircuitBreaker(nil, 3, time.Second); err == nil {
		t.Error("expected error for nil inner")
	}
	if _, err := NewCircuitBreaker(inner, 0, time.Second); err == nil {
		t.Error("expected error for threshold=0")
	}
	if _, err := NewCircuitBreaker(inner, 3, 0); err == nil {
		t.Error("expected error for zero cooldown")
	}
}

func TestCircuitBreaker_ClosedOnSuccess(t *testing.T) {
	inner := &stubLoader{result: map[string]string{"K": "V"}}
	cb, _ := NewCircuitBreaker(inner, 2, time.Second)

	got, err := cb.Load(context.Background(), "secret/app")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got["K"] != "V" {
		t.Errorf("expected V, got %s", got["K"])
	}
	if cb.State() != "closed" {
		t.Errorf("expected closed, got %s", cb.State())
	}
}

func TestCircuitBreaker_OpensAfterThreshold(t *testing.T) {
	inner := &stubLoader{err: errors.New("vault unavailable")}
	cb, _ := NewCircuitBreaker(inner, 3, time.Minute)

	for i := 0; i < 3; i++ {
		_, _ = cb.Load(context.Background(), "secret/app")
	}
	if cb.State() != "open" {
		t.Errorf("expected open after threshold, got %s", cb.State())
	}
	// Next call should return ErrCircuitOpen without calling inner.
	callsBefore := inner.calls
	_, err := cb.Load(context.Background(), "secret/app")
	if !errors.Is(err, ErrCircuitOpen) {
		t.Errorf("expected ErrCircuitOpen, got %v", err)
	}
	if inner.calls != callsBefore {
		t.Error("inner should not be called when circuit is open")
	}
}

func TestCircuitBreaker_RecoveryAfterCooldown(t *testing.T) {
	inner := &stubLoader{err: errors.New("fail")}
	cb, _ := NewCircuitBreaker(inner, 1, 10*time.Millisecond)

	_, _ = cb.Load(context.Background(), "secret/app")
	if cb.State() != "open" {
		t.Fatalf("expected open, got %s", cb.State())
	}

	time.Sleep(20 * time.Millisecond)

	// Switch inner to succeed.
	inner.err = nil
	inner.result = map[string]string{"X": "1"}

	got, err := cb.Load(context.Background(), "secret/app")
	if err != nil {
		t.Fatalf("unexpected error after cooldown: %v", err)
	}
	if got["X"] != "1" {
		t.Errorf("expected 1, got %s", got["X"])
	}
	if cb.State() != "closed" {
		t.Errorf("expected closed after recovery, got %s", cb.State())
	}
}

func TestCircuitBreaker_HalfOpenFailReopens(t *testing.T) {
	inner := &stubLoader{err: errors.New("fail")}
	cb, _ := NewCircuitBreaker(inner, 1, 10*time.Millisecond)

	_, _ = cb.Load(context.Background(), "secret/app")
	time.Sleep(20 * time.Millisecond)

	// Still failing in half-open → should reopen.
	_, err := cb.Load(context.Background(), "secret/app")
	if err == nil {
		t.Fatal("expected error in half-open with failing inner")
	}
	if cb.State() != "open" {
		t.Errorf("expected open after half-open failure, got %s", cb.State())
	}
}
