package sync

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestHashStage_PassesThroughSecrets(t *testing.T) {
	input := map[string]string{"FOO": "bar", "BAZ": "qux"}
	stage := HashStage(WithHashOutput(&bytes.Buffer{}))

	got, err := stage(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != len(input) {
		t.Errorf("expected %d secrets, got %d", len(input), len(got))
	}
	for k, v := range input {
		if got[k] != v {
			t.Errorf("key %q: want %q, got %q", k, v, got[k])
		}
	}
}

func TestHashStage_PrintsHashLine(t *testing.T) {
	var buf bytes.Buffer
	stage := HashStage(WithHashOutput(&buf))

	_, _ = stage(context.Background(), map[string]string{"K": "v"})

	output := buf.String()
	if !strings.Contains(output, "[vault-sync] secrets hash:") {
		t.Errorf("output missing hash prefix, got: %q", output)
	}
}

func TestHashStage_PrintsKeyCount(t *testing.T) {
	var buf bytes.Buffer
	stage := HashStage(WithHashOutput(&buf))

	_, _ = stage(context.Background(), map[string]string{"A": "1", "B": "2", "C": "3"})

	if !strings.Contains(buf.String(), "3 keys") {
		t.Errorf("expected '3 keys' in output, got: %q", buf.String())
	}
}

func TestHashStage_EmptySecrets_DoesNotError(t *testing.T) {
	var buf bytes.Buffer
	stage := HashStage(WithHashOutput(&buf))

	got, err := stage(context.Background(), map[string]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty map, got %d keys", len(got))
	}
	if !strings.Contains(buf.String(), "0 keys") {
		t.Errorf("expected '0 keys' in output, got: %q", buf.String())
	}
}

func TestHashStage_DefaultsToStdout(t *testing.T) {
	// Ensure construction without options does not panic.
	stage := HashStage()
	if stage == nil {
		t.Fatal("expected non-nil stage function")
	}
}
