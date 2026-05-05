package sync

import (
	"testing"
	"time"
)

func TestWatcherConfig_Validate_Disabled(t *testing.T) {
	c := WatcherConfig{Enabled: false}
	if err := c.Validate(); err != nil {
		t.Errorf("expected no error for disabled watcher, got %v", err)
	}
}

func TestWatcherConfig_Validate_MissingPaths(t *testing.T) {
	c := WatcherConfig{Enabled: true, Interval: 10 * time.Second}
	if err := c.Validate(); err == nil {
		t.Error("expected error for missing paths")
	}
}

func TestWatcherConfig_Validate_ShortInterval(t *testing.T) {
	c := WatcherConfig{Enabled: true, Interval: 500 * time.Millisecond, Paths: []string{"secret/app"}}
	if err := c.Validate(); err == nil {
		t.Error("expected error for interval < 1s")
	}
}

func TestWatcherConfig_Validate_Valid(t *testing.T) {
	c := WatcherConfig{Enabled: true, Interval: 30 * time.Second, Paths: []string{"secret/app"}}
	if err := c.Validate(); err != nil {
		t.Errorf("expected no error for valid config, got %v", err)
	}
}

func TestWatcherConfig_ToWatchOptions(t *testing.T) {
	called := false
	c := WatcherConfig{Interval: 15 * time.Second}
	opts := c.ToWatchOptions(func(_ string, _ interface{}) { called = true })

	if opts.Interval != 15*time.Second {
		t.Errorf("expected 15s interval, got %v", opts.Interval)
	}
	opts.OnChange("", nil)
	if !called {
		t.Error("expected OnChange to be wired to the provided callback")
	}
}
