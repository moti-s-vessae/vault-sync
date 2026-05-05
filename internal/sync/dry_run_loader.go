package sync

import (
	"context"
	"fmt"
	"io"
	"os"
)

// SecretsLoader is the interface satisfied by any loader the syncer uses.
type SecretsLoader interface {
	Load(ctx context.Context, path string) (map[string]string, error)
}

// DryRunLoader wraps a SecretsLoader and logs each Load call without
// forwarding to downstream writers. It is useful for --dry-run mode.
type DryRunLoader struct {
	inner SecretsLoader
	out   io.Writer
}

// NewDryRunLoader wraps inner, reporting Load calls to out.
// If out is nil, os.Stdout is used.
func NewDryRunLoader(inner SecretsLoader, out io.Writer) (*DryRunLoader, error) {
	if inner == nil {
		return nil, fmt.Errorf("dry-run loader: inner loader must not be nil")
	}
	if out == nil {
		out = os.Stdout
	}
	return &DryRunLoader{inner: inner, out: out}, nil
}

// Load delegates to the inner loader and logs the result count.
func (d *DryRunLoader) Load(ctx context.Context, path string) (map[string]string, error) {
	secrets, err := d.inner.Load(ctx, path)
	if err != nil {
		fmt.Fprintf(d.out, "[dry-run] load %s: error: %v\n", path, err)
		return nil, err
	}
	fmt.Fprintf(d.out, "[dry-run] load %s: fetched %d secret(s)\n", path, len(secrets))
	return secrets, nil
}
