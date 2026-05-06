package sync

import (
	"bytes"
	"strings"
	"testing"
)

func TestExportStage_SkipsWhenNoFormat(t *testing.T) {
	stage := ExportStage(ExportStageOptions{})
	secrets := map[string]string{"KEY": "val"}

	out, err := stage(secrets)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out["KEY"] != "val" {
		t.Errorf("secrets should pass through unchanged")
	}
}

func TestExportStage_InvalidFormat_ReturnsError(t *testing.T) {
	var buf bytes.Buffer
	stage := ExportStage(ExportStageOptions{Format: "toml", Out: &buf})

	_, err := stage(map[string]string{"K": "v"})
	if err == nil {
		t.Fatal("expected error for invalid format")
	}
}

func TestExportStage_WritesEnvFormat(t *testing.T) {
	var buf bytes.Buffer
	stage := ExportStage(ExportStageOptions{Format: FormatEnv, Out: &buf})

	secrets := map[string]string{"TOKEN": "abc123"}
	out, err := stage(secrets)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if out["TOKEN"] != "abc123" {
		t.Error("stage should not modify secrets")
	}
	if !strings.Contains(buf.String(), "TOKEN=abc123") {
		t.Errorf("expected TOKEN in output, got: %s", buf.String())
	}
}

func TestExportStage_WritesJSONFormat(t *testing.T) {
	var buf bytes.Buffer
	stage := ExportStage(ExportStageOptions{Format: FormatJSON, Out: &buf})

	secrets := map[string]string{"SECRET": "value"}
	_, err := stage(secrets)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(buf.String(), `"SECRET"`) {
		t.Errorf("expected JSON output, got: %s", buf.String())
	}
}

func TestExportStage_PassesThroughSecrets(t *testing.T) {
	var buf bytes.Buffer
	stage := ExportStage(ExportStageOptions{Format: FormatDotenv, Out: &buf})

	secrets := map[string]string{"A": "1", "B": "2"}
	out, err := stage(secrets)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out) != 2 {
		t.Errorf("expected 2 secrets, got %d", len(out))
	}
}
