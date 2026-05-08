package sync

import (
	"context"
	"testing"
)

type staticAnnotateLoader struct {
	data map[string]string
}

func (l *staticAnnotateLoader) Load(_ context.Context, _ string) (map[string]string, error) {
	out := make(map[string]string, len(l.data))
	for k, v := range l.data {
		out[k] = v
	}
	return out, nil
}

func TestAnnotateStage_Integration_WithPipeline(t *testing.T) {
	loader := &staticAnnotateLoader{
		data: map[string]string{
			"DB_PASSWORD": "s3cr3t",
			"APP_ENV":     "production",
		},
	}

	annotator, err := NewSecretAnnotator([]AnnotationRule{
		{Pattern: "^DB_", TagKey: "sensitivity", TagValue: "high"},
	})
	if err != nil {
		t.Fatalf("failed to create annotator: %v", err)
	}

	p, err := NewPipeline(loader, AnnotateStage(annotator))
	if err != nil {
		t.Fatalf("failed to create pipeline: %v", err)
	}

	result, err := p.Run(context.Background(), "secret/data/app")
	if err != nil {
		t.Fatalf("pipeline run failed: %v", err)
	}

	if result["DB_PASSWORD"] != "s3cr3t" {
		t.Errorf("expected DB_PASSWORD=s3cr3t, got %q", result["DB_PASSWORD"])
	}
	if result["DB_PASSWORD__sensitivity"] != "high" {
		t.Errorf("expected annotation DB_PASSWORD__sensitivity=high, got %q", result["DB_PASSWORD__sensitivity"])
	}
	if _, ok := result["APP_ENV__sensitivity"]; ok {
		t.Error("APP_ENV should not have sensitivity annotation")
	}
}

func TestAnnotateConfig_Validate_Disabled(t *testing.T) {
	cfg := &AnnotateConfig{Enabled: false}
	if err := cfg.Validate(); err != nil {
		t.Errorf("disabled config should be valid: %v", err)
	}
	a, err := cfg.ToAnnotator()
	if err != nil || a != nil {
		t.Error("disabled config should return nil annotator")
	}
}

func TestAnnotateConfig_Validate_MissingRules(t *testing.T) {
	cfg := &AnnotateConfig{Enabled: true, Rules: nil}
	if err := cfg.Validate(); err == nil {
		t.Error("expected error when enabled with no rules")
	}
}

func TestAnnotateConfig_ToAnnotator_Valid(t *testing.T) {
	cfg := &AnnotateConfig{
		Enabled: true,
		Rules:   []AnnotationRule{{Pattern: ".*", TagKey: "managed", TagValue: "true"}},
	}
	a, err := cfg.ToAnnotator()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a == nil {
		t.Fatal("expected non-nil annotator")
	}
}
