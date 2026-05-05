package sync

import (
	"context"
	"errors"
	"testing"
	"time"
)

type mockLoader struct {
	result map[string]string
	err    error
	calls  int
}

func (m *mockLoader) Load(_ context.Context, _ string) (map[string]string, error) {
	m.calls++
	return m.result, m.err
}

func TestNewRateLimitedLoader_NilInner(t *testing.T) {
	_, err := NewRateLimitedLoader(nil, 10)
	if err == nil {
		t.Fatal("expected error for nil inner loader")
	}
}

func TestNewRateLimitedLoader_InvalidRPS(t *testing.T) {
	_, err := NewRateLimitedLoader(&mockLoader{}, -1)
	if err == nil {
		t.Fatal("expected error for invalid rps")
	}
}

func TestRateLimitedLoader_Load_Success(t *testing.T) {
	inner := &mockLoader{result: map[string]string{"KEY": "val"}}
	rl, err := NewRateLimitedLoader(inner, 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, err := rl.Load(context.Background(), "secret/app")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got["KEY"] != "val" {
		t.Errorf("expected KEY=val, got %v", got)
	}
	if inner.calls != 1 {
		t.Errorf("expected 1 call, got %d", inner.calls)
	}
}

func TestRateLimitedLoader_Load_InnerError(t *testing.T) {
	inner := &mockLoader{err: errors.New("vault unavailable")}
	rl, _ := NewRateLimitedLoader(inner, 100)
	_, err := rl.Load(context.Background(), "secret/app")
	if err == nil {
		t.Fatal("expected error from inner loader")
	}
}

func TestRateLimitedLoader_Load_ContextCancelled(t *testing.T) {
	inner := &mockLoader{result: map[string]string{}}
	// very low rps so second call blocks
	rl, _ := NewRateLimitedLoader(inner, 0.01)
	// consume the initial token
	_, _ = rl.Load(context.Background(), "secret/app")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	_, err := rl.Load(ctx, "secret/app")
	if err == nil {
		t.Fatal("expected context deadline error")
	}
	if inner.calls != 1 {
		t.Errorf("inner should only be called once, got %d", inner.calls)
	}
}
