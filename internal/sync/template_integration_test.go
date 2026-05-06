package sync

import (
	"strings"
	"testing"
)

// TestTemplateStage_Integration_WithPipeline verifies that TemplateStage
// integrates correctly as a pipeline stage, producing a derived secret
// from upstream values alongside the original secrets.
func TestTemplateStage_Integration_WithPipeline(t *testing.T) {
	loader := &mockLoader{
		secrets: map[string]string{
			"DB_HOST": "db.internal",
			"DB_PORT": "5432",
			"DB_NAME": "appdb",
			"DB_USER": "appuser",
			"DB_PASS": "hunter2",
		},
	}

	const dsn = `postgres://{{ index . "DB_USER" }}:{{ index . "DB_PASS" }}@{{ index . "DB_HOST" }}:{{ index . "DB_PORT" }}/{{ index . "DB_NAME" }}`

	p, err := NewPipeline(loader, TemplateStage(dsn, "DATABASE_URL"))
	if err != nil {
		t.Fatalf("NewPipeline error: %v", err)
	}

	result, err := p.Run(t.Context())
	if err != nil {
		t.Fatalf("pipeline Run error: %v", err)
	}

	want := "postgres://appuser:hunter2@db.internal:5432/appdb"
	if result["DATABASE_URL"] != want {
		t.Errorf("DATABASE_URL = %q, want %q", result["DATABASE_URL"], want)
	}

	// Original secrets must still be present.
	for k, v := range loader.secrets {
		if result[k] != v {
			t.Errorf("secret %q = %q, want %q", k, result[k], v)
		}
	}
}

func TestTemplateStage_Integration_MissingKey_PropagatesError(t *testing.T) {
	loader := &mockLoader{
		secrets: map[string]string{
			"ONLY_KEY": "value",
		},
	}

	p, err := NewPipeline(loader, TemplateStage(`{{ index . "MISSING" }}`, "OUT"))
	if err != nil {
		t.Fatalf("NewPipeline error: %v", err)
	}

	_, err = p.Run(t.Context())
	if err == nil {
		t.Fatal("expected error when template references missing key")
	}
	if !strings.Contains(err.Error(), "render failed") {
		t.Errorf("unexpected error: %v", err)
	}
}
