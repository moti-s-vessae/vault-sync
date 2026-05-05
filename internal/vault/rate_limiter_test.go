package vault

import (
	"context"
	"testing"
	"time"
)

func TestNewRateLimiter_InvalidRPS(t *testing.T) {
	_, err := NewRateLimiter(0)
	if err == nil {
		t.Fatal("expected error for rps=0")
	}
	_, err = NewRateLimiter(-1)
	if err == nil {
		t.Fatal("expected error for rps=-1")
	}
}

func TestNewRateLimiter_Valid(t *testing.T) {
	rl, err := NewRateLimiter(10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rl == nil {
		t.Fatal("expected non-nil RateLimiter")
	}
}

func TestRateLimiter_Wait_ImmediateFirstToken(t *testing.T) {
	rl, _ := NewRateLimiter(10)
	ctx := context.Background()
	start := time.Now()
	if err := rl.Wait(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if elapsed := time.Since(start); elapsed > 50*time.Millisecond {
		t.Errorf("first token should be immediate, took %v", elapsed)
	}
}

func TestRateLimiter_Wait_ContextCancelled(t *testing.T) {
	// 0.1 rps means ~10s between tokens — context should cancel first
	rl, _ := NewRateLimiter(0.1)
	// drain initial token
	_ = rl.Wait(context.Background())

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := rl.Wait(ctx)
	if err == nil {
		t.Fatal("expected context cancellation error")
	}
}

func TestRateLimiter_Available(t *testing.T) {
	rl, _ := NewRateLimiter(5)
	av := rl.Available()
	if av < 4.9 || av > 5.1 {
		t.Errorf("expected ~5 available tokens, got %f", av)
	}
	_ = rl.Wait(context.Background())
	av2 := rl.Available()
	if av2 >= av {
		t.Errorf("expected fewer tokens after Wait, got %f", av2)
	}
}

func TestRateLimiter_Wait_MultipleTokens(t *testing.T) {
	rl, _ := NewRateLimiter(100) // high rps for fast test
	ctx := context.Background()
	for i := 0; i < 5; i++ {
		if err := rl.Wait(ctx); err != nil {
			t.Fatalf("unexpected error on iteration %d: %v", i, err)
		}
	}
}
