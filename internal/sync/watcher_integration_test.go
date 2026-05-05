package sync

import (
	"context"
	"sync"
	"testing"
	"time"
)

// TestWatcher_Integration_DetectsRotation simulates a secret rotation while
// the watcher is running and asserts the onChange callback fires exactly once.
func TestWatcher_Integration_DetectsRotation(t *testing.T) {
	var mu sync.Mutex
	changes := 0

	loader := &mockWatchLoader{
		response: map[string]string{"DB_PASSWORD": "initial"},
	}

	cfg := WatcherConfig{
		Enabled:  true,
		Interval: 20 * time.Millisecond,
		Paths:    []string{"secret/db"},
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("invalid config: %v", err)
	}

	w, err := NewSecretWatcher(loader, cfg.Paths, cfg.ToWatchOptions(func(path string, _ interface{}) {
		mu.Lock()
		changes++
		mu.Unlock()
	}))
	if err != nil {
		t.Fatalf("NewSecretWatcher: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() { errCh <- w.Run(ctx) }()

	// Allow the initial poll to complete.
	time.Sleep(30 * time.Millisecond)

	// Rotate the secret.
	loader.mu.Lock()
	loader.response = map[string]string{"DB_PASSWORD": "rotated"}
	loader.mu.Unlock()

	// Allow at least one more tick to detect the change.
	time.Sleep(50 * time.Millisecond)
	cancel()
	<-errCh

	mu.Lock()
	defer mu.Unlock()
	if changes != 1 {
		t.Errorf("expected exactly 1 onChange call, got %d", changes)
	}
}
