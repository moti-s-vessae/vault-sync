package sync

import (
	"context"
	"testing"
	"time"
)

func TestDefaultBackoffConfig_IsValid(t *testing.T) {
	cfg := DefaultBackoffConfig()
	if err := cfg.Validate(); err != nil {
		t.Fatalf("DefaultBackoffConfig should be valid, got: %v", err)
	}
}

func TestBackoffConfig_Validate_Errors(t *testing.T) {
	tests := []struct {
		name string
		cfg  BackoffConfig
	}{
		{"zero initial", BackoffConfig{InitialInterval: 0, MaxInterval: time.Second, JitterFraction: 0}},
		{"max < initial", BackoffConfig{InitialInterval: time.Second, MaxInterval: time.Millisecond, JitterFraction: 0}},
		{"jitter > 1", BackoffConfig{InitialInterval: time.Millisecond, MaxInterval: time.Second, JitterFraction: 1.5}},
		{"jitter < 0", BackoffConfig{InitialInterval: time.Millisecond, MaxInterval: time.Second, JitterFraction: -0.1}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.cfg.Validate(); err == nil {
				t.Errorf("expected validation error for config %+v", tc.cfg)
			}
		})
	}
}

func TestBackoff_GrowsExponentially(t *testing.T) {
	cfg := BackoffConfig{
		InitialInterval: 100 * time.Millisecond,
		Multiplier:      2.0,
		MaxInterval:     10 * time.Second,
		JitterFraction:  0, // no jitter for deterministic test
	}
	expected := []time.Duration{
		100 * time.Millisecond,
		200 * time.Millisecond,
		400 * time.Millisecond,
		800 * time.Millisecond,
	}
	for i, want := range expected {
		got := cfg.Backoff(i)
		if got != want {
			t.Errorf("attempt %d: want %v, got %v", i, want, got)
		}
	}
}

func TestBackoff_CapsAtMaxInterval(t *testing.T) {
	cfg := BackoffConfig{
		InitialInterval: 100 * time.Millisecond,
		Multiplier:      2.0,
		MaxInterval:     500 * time.Millisecond,
		JitterFraction:  0,
	}
	for attempt := 10; attempt < 20; attempt++ {
		got := cfg.Backoff(attempt)
		if got > cfg.MaxInterval {
			t.Errorf("attempt %d: got %v, exceeds MaxInterval %v", attempt, got, cfg.MaxInterval)
		}
	}
}

func TestBackoff_Sleep_ContextCancelled(t *testing.T) {
	cfg := BackoffConfig{
		InitialInterval: 10 * time.Second,
		Multiplier:      2.0,
		MaxInterval:     30 * time.Second,
		JitterFraction:  0,
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	start := time.Now()
	err := cfg.Sleep(ctx, 0)
	if err == nil {
		t.Fatal("expected context cancellation error")
	}
	if time.Since(start) > time.Second {
		t.Error("Sleep did not return promptly after context cancellation")
	}
}

func TestBackoff_Sleep_Completes(t *testing.T) {
	cfg := BackoffConfig{
		InitialInterval: 10 * time.Millisecond,
		Multiplier:      2.0,
		MaxInterval:     time.Second,
		JitterFraction:  0,
	}
	ctx := context.Background()
	if err := cfg.Sleep(ctx, 0); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
