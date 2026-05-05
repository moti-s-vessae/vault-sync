package sync_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/your-org/vault-sync/internal/sync"
)

func TestDryRunWriter_PrintsPath(t *testing.T) {
	var buf bytes.Buffer
	w := sync.NewDryRunWriter(&buf)

	err := w.WriteEnvFile(".env", map[string]string{"FOO": "bar"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, ".env") {
		t.Errorf("expected output to contain path '.env', got: %s", out)
	}
}

func TestDryRunWriter_PrintsKeyCount(t *testing.T) {
	var buf bytes.Buffer
	w := sync.NewDryRunWriter(&buf)

	secrets := map[string]string{"A": "1", "B": "2", "C": "3"}
	_ = w.WriteEnvFile(".env", secrets)

	out := buf.String()
	if !strings.Contains(out, "3 secret(s)") {
		t.Errorf("expected '3 secret(s)' in output, got: %s", out)
	}
}

func TestDryRunWriter_MasksValues(t *testing.T) {
	var buf bytes.Buffer
	w := sync.NewDryRunWriter(&buf)

	_ = w.WriteEnvFile(".env", map[string]string{"SECRET_KEY": "super-secret"})

	out := buf.String()
	if strings.Contains(out, "super-secret") {
		t.Errorf("dry-run output must not contain secret values")
	}
	if !strings.Contains(out, "SECRET_KEY=***") {
		t.Errorf("expected masked key in output, got: %s", out)
	}
}

func TestDryRunWriter_SortedKeys(t *testing.T) {
	var buf bytes.Buffer
	w := sync.NewDryRunWriter(&buf)

	_ = w.WriteEnvFile(".env", map[string]string{"ZZZ": "z", "AAA": "a", "MMM": "m"})

	out := buf.String()
	idxA := strings.Index(out, "AAA")
	idxM := strings.Index(out, "MMM")
	idxZ := strings.Index(out, "ZZZ")
	if !(idxA < idxM && idxM < idxZ) {
		t.Errorf("expected sorted output, got: %s", out)
	}
}

func TestDryRunWriter_DefaultsToStdout(t *testing.T) {
	// Should not panic when out is nil.
	w := sync.NewDryRunWriter(nil)
	if w == nil {
		t.Fatal("expected non-nil writer")
	}
}

func TestDryRunWriter_EmptySecrets(t *testing.T) {
	var buf bytes.Buffer
	w := sync.NewDryRunWriter(&buf)

	err := w.WriteEnvFile(".env", map[string]string{})
	if err != nil {
		t.Fatalf("unexpected error on empty secrets: %v", err)
	}

	if !strings.Contains(buf.String(), "0 secret(s)") {
		t.Errorf("expected '0 secret(s)' in output, got: %s", buf.String())
	}
}
