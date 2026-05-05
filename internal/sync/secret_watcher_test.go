package sync

import (
	"context"
	"sync"
	"testing"
	"time"
)

// mockWatchLoader records Load calls and returns preset responses.
type mockWatchLoader struct {
	mu       sync.Mutex
	calls    int
	response map[string]string
	err      error
}

func (m *mockWatchLoader) Load(_ context.Context, _ string) (map[string]string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls++
	return m.response, m.err
}

func TestNewSecretWatcher_NilLoader(t *testing.T) {
	_, err := NewSecretWatcher(nil, []string{"secret/app"}, WatchOptions{})
	if err == nil {
		t.Fatal("expected error for nil loader")
	}
}

func TestNewSecretWatcher_DefaultInterval(t *testing.T) {
	l := &mockWatchLoader{response: map[string]string{}}
	w, err := NewSecretWatcher(l, []string{"secret/app"}, WatchOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if w.opts.Interval != 30*time.Second {
		t.Errorf("expected 30s default interval, got %v", w.opts.Interval)
	}
}

func TestSecretWatcher_InvokesOnChange(t *testing.T) {
	call := 0
	loader := &mockWatchLoader{response: map[string]string{"KEY": "v1"}}

	w, _ := NewSecretWatcher(loader, []string{"secret/app"}, WatchOptions{
		Interval: 20 * time.Millisecond,
		OnChange: func(_ string, _ interface{}) { call++ },
	})

	// Seed initial snapshot without triggering onChange.
	ctx, cancel := context.WithTimeout(context.Background(), 80*time.Millisecond)
	defer cancel()

	w.poll(ctx) // first poll — no onChange yet
	if call != 0 {
		t.Fatalf("expected 0 onChange calls after first poll, got %d", call)
	}

	// Change the secret so next poll detects a diff.
	loader.mu.Lock()
	loader.response = map[string]string{"KEY": "v2"}
	loader.mu.Unlock()

	w.poll(ctx)
	if call != 1 {
		t.Fatalf("expected 1 onChange call after change, got %d", call)
	}
}

func TestSecretWatcher_NoChangeNoCallback(t *testing.T) {
	call := 0
	loader := &mockWatchLoader{response: map[string]string{"KEY": "same"}}

	w, _ := NewSecretWatcher(loader, []string{"secret/app"}, WatchOptions{
		Interval: 20 * time.Millisecond,
		OnChange: func(_ string, _ interface{}) { call++ },
	})

	ctx := context.Background()
	w.poll(ctx)
	w.poll(ctx)
	w.poll(ctx)

	if call != 0 {
		t.Errorf("expected no onChange calls when secrets unchanged, got %d", call)
	}
}

func TestSecretWatcher_RunCancels(t *testing.T) {
	loader := &mockWatchLoader{response: map[string]string{}}
	w, _ := NewSecretWatcher(loader, []string{"secret/app"}, WatchOptions{
		Interval: 10 * time.Millisecond,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := w.Run(ctx)
	if err != context.DeadlineExceeded {
		t.Errorf("expected DeadlineExceeded, got %v", err)
	}
}
