package sync_test

import (
	"context"
	"sync"
	"testing"
	"time"

	vsync "github.com/your-org/vault-sync/internal/sync"
)

type rotatingFilterLoader struct {
	mu       sync.Mutex
	current  map[string]string
}

func (r *rotatingFilterLoader) Load(_ string) (map[string]string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make(map[string]string, len(r.current))
	for k, v := range r.current {
		out[k] = v
	}
	return out, nil
}

func (r *rotatingFilterLoader) set(m map[string]string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.current = m
}

func TestWatcher_Integration_FilterStageDetectsChange(t *testing.T) {
	loader := &rotatingFilterLoader{
		current: map[string]string{
			"DB_HOST": "localhost",
			"DB_PASS": "old",
		},
	}

	rules := []vsync.FilterRule{{Pattern: "^DB_"}}

	changed := make(chan map[string]string, 1)

	watcher, err := vsync.NewSecretWatcher(loader, vsync.WatchOptions{
		Paths:    []string{"secret/app"},
		Interval: 50 * time.Millisecond,
		OnChange: func(path string, secrets map[string]string) {
			filtered := vsync.KeyFilterStage(rules)
			result, _ := filtered(secrets)
			changed <- result
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	go watcher.Start(ctx)

	time.Sleep(80 * time.Millisecond)
	loader.set(map[string]string{
		"DB_HOST": "localhost",
		"DB_PASS": "rotated",
	})

	select {
	case result := <-changed:
		if result["DB_PASS"] != "rotated" {
			t.Errorf("expected rotated password, got %q", result["DB_PASS"])
		}
	case <-ctx.Done():
		t.Fatal("timed out waiting for secret change callback")
	}
}
