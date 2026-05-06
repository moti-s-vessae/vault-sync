package sync

import (
	"bytes"
	"strings"
	"testing"
)

type staticLoader struct {
	secrets map[string]string
}

func (s *staticLoader) Load(path string) (map[string]string, error) {
	return s.secrets, nil
}

func TestExportStage_Integration_WithPipeline(t *testing.T) {
	loader := &staticLoader{
		secrets: map[string]string{
			"APP_SECRET": "topsecret",
			"DB_URL":     "postgres://localhost/db",
		},
	}

	var buf bytes.Buffer
	p, err := NewPipeline(loader,
		ExportStage(ExportStageOptions{Format: FormatEnv, Out: &buf}),
	)
	if err != nil {
		t.Fatalf("NewPipeline error: %v", err)
	}

	result, err := p.Run("secret/app")
	if err != nil {
		t.Fatalf("pipeline Run error: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("expected 2 secrets, got %d", len(result))
	}

	output := buf.String()
	if !strings.Contains(output, "APP_SECRET=topsecret") {
		t.Errorf("expected APP_SECRET in export output, got: %s", output)
	}
	if !strings.Contains(output, "DB_URL=") {
		t.Errorf("expected DB_URL in export output, got: %s", output)
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 lines in output, got %d", len(lines))
	}
	if lines[0] > lines[1] {
		t.Errorf("expected sorted output, got: %v", lines)
	}
}
