package sync

import (
	"context"
	"log"
	"time"
)

// WatchOptions configures the secret watcher behaviour.
type WatchOptions struct {
	Interval time.Duration
	OnChange func(path string, diff interface{})
	Logger   *log.Logger
}

// SecretWatcher polls a secrets loader at a fixed interval and invokes
// a callback whenever the returned secrets differ from the previous snapshot.
type SecretWatcher struct {
	loader   SecretsLoader
	paths    []string
	opts     WatchOptions
	snapshot map[string]map[string]string
}

// SecretsLoader is the minimal interface required by SecretWatcher.
type SecretsLoader interface {
	Load(ctx context.Context, path string) (map[string]string, error)
}

// NewSecretWatcher creates a SecretWatcher that monitors the given paths.
func NewSecretWatcher(loader SecretsLoader, paths []string, opts WatchOptions) (*SecretWatcher, error) {
	if loader == nil {
		return nil, ErrNilLoader
	}
	if opts.Interval <= 0 {
		opts.Interval = 30 * time.Second
	}
	if opts.Logger == nil {
		opts.Logger = log.Default()
	}
	return &SecretWatcher{
		loader:   loader,
		paths:    paths,
		opts:     opts,
		snapshot: make(map[string]map[string]string),
	}, nil
}

// Run starts the watch loop and blocks until ctx is cancelled.
func (w *SecretWatcher) Run(ctx context.Context) error {
	ticker := time.NewTicker(w.opts.Interval)
	defer ticker.Stop()

	w.poll(ctx) // initial poll

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			w.poll(ctx)
		}
	}
}

func (w *SecretWatcher) poll(ctx context.Context) {
	for _, path := range w.paths {
		secrets, err := w.loader.Load(ctx, path)
		if err != nil {
			w.opts.Logger.Printf("watcher: error loading %q: %v", path, err)
			continue
		}
		prev, seen := w.snapshot[path]
		if !seen || hasChanged(prev, secrets) {
			w.snapshot[path] = copyMap(secrets)
			if seen && w.opts.OnChange != nil {
				w.opts.OnChange(path, secrets)
			}
		}
	}
}

func hasChanged(prev, next map[string]string) bool {
	if len(prev) != len(next) {
		return true
	}
	for k, v := range next {
		if prev[k] != v {
			return true
		}
	}
	return false
}

func copyMap(m map[string]string) map[string]string {
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}
