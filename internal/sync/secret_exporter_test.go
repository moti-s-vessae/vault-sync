package sync

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestNewSecretExporter_InvalidFormat(t *testing.T) {
	_, err := NewSecretExporter("xml", nil)
	if err == nil {
		t.Fatal("expected error for unsupported format")
	}
}

func TestNewSecretExporter_DefaultsToStdout(t *testing.T) {
	e, err := NewSecretExporter(FormatEnv, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if e.out == nil {
		t.Fatal("expected non-nil writer")
	}
}

func TestSecretExporter_ExportEnv_SortedOutput(t *testing.T) {
	var buf bytes.Buffer
	e, _ := NewSecretExporter(FormatEnv, &buf)

	secrets := map[string]string{
		"Z_KEY": "last",
		"A_KEY": "first",
		"M_KEY": "middle",
	}
	if err := e.Export(secrets); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}
	if !strings.HasPrefix(lines[0], "A_KEY=") {
		t.Errorf("expected first line to be A_KEY, got %s", lines[0])
	}
	if !strings.HasPrefix(lines[2], "Z_KEY=") {
		t.Errorf("expected last line to be Z_KEY, got %s", lines[2])
	}
}

func TestSecretExporter_ExportEnv_QuotesSpecialValues(t *testing.T) {
	var buf bytes.Buffer
	e, _ := NewSecretExporter(FormatDotenv, &buf)

	secrets := map[string]string{"KEY": "hello world"}
	if err := e.Export(secrets); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, `"hello world"`) {
		t.Errorf("expected quoted value, got: %s", output)
	}
}

func TestSecretExporter_ExportJSON_ValidJSON(t *testing.T) {
	var buf bytes.Buffer
	e, _ := NewSecretExporter(FormatJSON, &buf)

	secrets := map[string]string{"DB_PASS": "secret123", "API_KEY": "abc"}
	if err := e.Export(secrets); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed map[string]string
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if parsed["DB_PASS"] != "secret123" {
		t.Errorf("unexpected value: %s", parsed["DB_PASS"])
	}
}

func TestSecretExporter_ExportEnv_PlainValues(t *testing.T) {
	var buf bytes.Buffer
	e, _ := NewSecretExporter(FormatEnv, &buf)

	secrets := map[string]string{"SIMPLE": "value"}
	if err := e.Export(secrets); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(buf.String(), "SIMPLE=value") {
		t.Errorf("expected plain assignment, got: %s", buf.String())
	}
}
