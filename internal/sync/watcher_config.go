package sync

import (
	"errors"
	"time"
)

// WatcherConfig holds configuration for the SecretWatcher loaded from the
// application config file.
type WatcherConfig struct {
	Enabled  bool          `yaml:"enabled"`
	Interval time.Duration `yaml:"interval"`
	Paths    []string      `yaml:"paths"`
}

// Validate returns an error if the WatcherConfig is semantically invalid.
func (c WatcherConfig) Validate() error {
	if !c.Enabled {
		return nil
	}
	if c.Interval < time.Second {
		return errors.New("watcher: interval must be at least 1s")
	}
	if len(c.Paths) == 0 {
		return errors.New("watcher: at least one path is required when enabled")
	}
	return nil
}

// ToWatchOptions converts WatcherConfig into WatchOptions.
func (c WatcherConfig) ToWatchOptions(onChange func(string, interface{})) WatchOptions {
	return WatchOptions{
		Interval: c.Interval,
		OnChange: onChange,
	}
}
