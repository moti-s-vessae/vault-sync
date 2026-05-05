package sync

import (
	"context"
	"testing"
	"time"
)

func TestRateLimitedLoader_Integration_ThrottlesRequests(t *testing.T) {
	const rps = 20.0
	const calls = 5

	inner := &mockLoader{result: map[string]string{"X": "1"}}
	rl, err := NewRateLimitedLoader(inner, rps)
	if err != nil {
		t.Fatalf("setup error: %v", err)
	}

	ctx := context.Background()
	start := time.Now()
	for i := 0; i < calls; i++ {
		if _, err := rl.Load(ctx, "secret/test"); err != nil {
			t.Fatalf("call %d failed: %v", i, err)
		}
	}
	elapsed := time.Since(start)

	// With rps=20 and 5 calls (first is free), we expect at least 4*(1/20)=200ms
	minExpected := time.Duration(float64(calls-1) / rps * float64(time.Second))
	if elapsed < minExpected/2 {
		t.Errorf("calls completed too fast (%v), expected at least %v", elapsed, minExpected/2)
	}
	if inner.calls != calls {
		t.Errorf("expected %d inner calls, got %d", calls, inner.calls)
	}
}
