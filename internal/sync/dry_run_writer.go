package sync

import (
	"fmt"
	"io"
	"os"
	"sort"
)

// DryRunWriter is an env file writer that prints what would be written
// without modifying any files on disk.
type DryRunWriter struct {
	out io.Writer
}

// NewDryRunWriter creates a DryRunWriter that reports planned writes to out.
// If out is nil, os.Stdout is used.
func NewDryRunWriter(out io.Writer) *DryRunWriter {
	if out == nil {
		out = os.Stdout
	}
	return &DryRunWriter{out: out}
}

// WriteEnvFile prints the secrets that would be written to path instead of
// actually writing the file. It satisfies the same interface as env.WriteEnvFile.
func (d *DryRunWriter) WriteEnvFile(path string, secrets map[string]string) error {
	keys := make([]string, 0, len(secrets))
	for k := range secrets {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	fmt.Fprintf(d.out, "[dry-run] would write %d secret(s) to %s:\n", len(secrets), path)
	for _, k := range keys {
		fmt.Fprintf(d.out, "  %s=***\n", k)
	}
	return nil
}
