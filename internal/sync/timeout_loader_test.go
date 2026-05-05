package sync

import (
	"context"
	"errors"
	"testing"
	"time"
)

// slowLoader simulates a loader that blocks until its context is cancelled.
type slowLoader struct {
	delay time.Duration
}

func (s *slowLoader) Load(ctx context.Context, _ string) (map[string]string, error) {
	select {
	case <-time.After(s.delay):
		return map[string]string{"KEY": "val"}, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

type fixedLoader struct {
	data map[string]string
	err  error
}

func (f *fixedLoader) Load(_ context.Context, _ string) (map[string]string, error) {
	return f.data, f.err
}

func TestNewTimeoutLoader_NilInner(t *testing.T) {
	_, err := NewTimeoutLoader(nil, time.Second)
	if err == nil {
		t.Fatal("expected error for nil inner loader")
	}
}

func TestNewTimeoutLoader_NonPositiveTimeout(t *testing.T) {
	_, err := NewTimeoutLoader(&fixedLoader{}, 0)
	if err == nil {
		t.Fatal("expected error for zero timeout")
	}
	_, err = NewTimeoutLoader(&fixedLoader{}, -time.Second)
	if err == nil {
		t.Fatal("expected error for negative timeout")
	}
}

func TestTimeoutLoader_Load_Success(t *testing.T) {
	expected := map[string]string{"FOO": "bar"}
	l, err := NewTimeoutLoader(&fixedLoader{data: expected}, time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, err := l.Load(context.Background(), "secret/app")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got["FOO"] != "bar" {
		t.Errorf("expected FOO=bar, got %v", got)
	}
}

func TestTimeoutLoader_Load_InnerError(t *testing.T) {
	sentinel := errors.New("vault unavailable")
	l, _ := NewTimeoutLoader(&fixedLoader{err: sentinel}, time.Second)
	_, err := l.Load(context.Background(), "secret/app")
	if !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel error, got %v", err)
	}
}

func TestTimeoutLoader_Load_TimesOut(t *testing.T) {
	l, err := NewTimeoutLoader(&slowLoader{delay: 200 * time.Millisecond}, 30*time.Millisecond)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = l.Load(context.Background(), "secret/app")
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
}

func TestTimeoutLoader_Load_ParentCancelled(t *testing.T) {
	l, _ := NewTimeoutLoader(&slowLoader{delay: time.Second}, 500*time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := l.Load(ctx, "secret/app")
	if err == nil {
		t.Fatal("expected error when parent context is already cancelled")
	}
}
