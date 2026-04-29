package sync

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/example/vault-sync/internal/vault"
)

// AuditEntry records a single sync event.
type AuditEntry struct {
	Timestamp time.Time
	Path      string
	Diff      []vault.SecretChange
}

// AuditLogger writes sync audit entries to a writer.
type AuditLogger struct {
	w io.Writer
}

// NewAuditLogger creates an AuditLogger writing to the given path.
// Pass an empty path to disable audit logging (writes to io.Discard).
func NewAuditLogger(path string) (*AuditLogger, error) {
	if path == "" {
		return &AuditLogger{w: io.Discard}, nil
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return nil, fmt.Errorf("audit: open log file: %w", err)
	}
	return &AuditLogger{w: f}, nil
}

// Log writes an audit entry for the given sync operation.
func (a *AuditLogger) Log(entry AuditEntry) error {
	ts := entry.Timestamp.UTC().Format(time.RFC3339)
	for _, ch := range entry.Diff {
		line := fmt.Sprintf("%s\t%s\t%s\t%s\n", ts, entry.Path, ch.Action, ch.Key)
		if _, err := fmt.Fprint(a.w, line); err != nil {
			return fmt.Errorf("audit: write entry: %w", err)
		}
	}
	if len(entry.Diff) == 0 {
		line := fmt.Sprintf("%s\t%s\tno-change\t-\n", ts, entry.Path)
		if _, err := fmt.Fprint(a.w, line); err != nil {
			return fmt.Errorf("audit: write entry: %w", err)
		}
	}
	return nil
}
