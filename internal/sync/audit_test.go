package sync

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/example/vault-sync/internal/vault"
)

func TestAuditLogger_LogChanges(t *testing.T) {
	var buf bytes.Buffer
	logger := &AuditLogger{w: &buf}

	entry := AuditEntry{
		Timestamp: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
		Path:      ".env",
		Diff: []vault.SecretChange{
			{Key: "DB_HOST", Action: "added"},
			{Key: "API_KEY", Action: "changed"},
		},
	}

	if err := logger.Log(entry); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "DB_HOST") {
		t.Errorf("expected DB_HOST in output, got: %s", out)
	}
	if !strings.Contains(out, "added") {
		t.Errorf("expected 'added' action in output, got: %s", out)
	}
	if !strings.Contains(out, "API_KEY") {
		t.Errorf("expected API_KEY in output, got: %s", out)
	}
	if !strings.Contains(out, "2024-01-15T10:00:00Z") {
		t.Errorf("expected timestamp in output, got: %s", out)
	}
}

func TestAuditLogger_LogNoChanges(t *testing.T) {
	var buf bytes.Buffer
	logger := &AuditLogger{w: &buf}

	entry := AuditEntry{
		Timestamp: time.Now(),
		Path:      ".env",
		Diff:      []vault.SecretChange{},
	}

	if err := logger.Log(entry); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "no-change") {
		t.Errorf("expected 'no-change' in output, got: %s", out)
	}
}

func TestNewAuditLogger_EmptyPath(t *testing.T) {
	logger, err := NewAuditLogger("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if logger == nil {
		t.Fatal("expected non-nil logger")
	}
	// logging to discard should not error
	err = logger.Log(AuditEntry{
		Timestamp: time.Now(),
		Path:      ".env",
		Diff:      []vault.SecretChange{{Key: "X", Action: "added"}},
	})
	if err != nil {
		t.Errorf("unexpected error logging to discard: %v", err)
	}
}

func TestNewAuditLogger_InvalidPath(t *testing.T) {
	_, err := NewAuditLogger("/nonexistent/dir/audit.log")
	if err == nil {
		t.Fatal("expected error for invalid path")
	}
}
