package sync

import (
	"context"
	"fmt"
	"io"
	"os"
)

// HashStageOption configures HashStage behaviour.
type HashStageOption func(*hashStageConfig)

type hashStageConfig struct {
	out io.Writer
}

// WithHashOutput redirects hash output to the given writer (default: os.Stdout).
func WithHashOutput(w io.Writer) HashStageOption {
	return func(c *hashStageConfig) { c.out = w }
}

// HashStage is a pipeline stage that computes and prints the SHA-256 hash of
// the current secret map. It passes secrets through unchanged.
func HashStage(opts ...HashStageOption) func(context.Context, map[string]string) (map[string]string, error) {
	cfg := &hashStageConfig{out: os.Stdout}
	for _, o := range opts {
		o(cfg)
	}

	hasher := NewSecretHasher()

	return func(_ context.Context, secrets map[string]string) (map[string]string, error) {
		hash := hasher.Hash(secrets)
		fmt.Fprintf(cfg.out, "[vault-sync] secrets hash: %s (%d keys)\n", hash, len(secrets))
		return secrets, nil
	}
}
