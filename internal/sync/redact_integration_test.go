package sync

import (
	"context"
	"testing"
)

type staticLoader struct {
	data map[string]string
}

func (s *staticLoader) Load(_ context.Context, _ string) (map[string]string, error) {
	out := make(map[string]string, len(s.data))
	for k, v := range s.data {
		out[k] = v
	}
	return out, nil
}

func TestRedactStage_Integration_WithPipeline(t *testing.T) {
	loader := &staticLoader{
		data: map[string]string{
			"DB_PASSWORD": "super-secret",
			"API_TOKEN":   "tok-abc123",
			"APP_ENV":     "production",
		},
	}

	rules := []RedactRule{
		{Pattern: "(?i)password", Replacement: "[REDACTED]"},
		{Pattern: "(?i)token", Replacement: "[REDACTED]"},
	}

	pipeline, err := NewPipeline(loader, RedactStage(rules))
	if err != nil {
		t.Fatalf("NewPipeline error: %v", err)
	}

	result, err := pipeline.Run(context.Background(), "secret/app")
	if err != nil {
		t.Fatalf("pipeline.Run error: %v", err)
	}

	if result["DB_PASSWORD"] != "[REDACTED]" {
		t.Errorf("DB_PASSWORD: expected [REDACTED], got %q", result["DB_PASSWORD"])
	}
	if result["API_TOKEN"] != "[REDACTED]" {
		t.Errorf("API_TOKEN: expected [REDACTED], got %q", result["API_TOKEN"])
	}
	if result["APP_ENV"] != "production" {
		t.Errorf("APP_ENV: expected production, got %q", result["APP_ENV"])
	}
}

func TestRedactStage_Integration_NoMatchKeepsAll(t *testing.T) {
	loader := &staticLoader{
		data: map[string]string{
			"HOST": "localhost",
			"PORT": "5432",
		},
	}

	pipeline, err := NewPipeline(loader, RedactStage([]RedactRule{
		{Pattern: "SECRET", Replacement: "[REDACTED]"},
	}))
	if err != nil {
		t.Fatalf("NewPipeline error: %v", err)
	}

	result, err := pipeline.Run(context.Background(), "secret/db")
	if err != nil {
		t.Fatalf("pipeline.Run error: %v", err)
	}

	if result["HOST"] != "localhost" || result["PORT"] != "5432" {
		t.Errorf("unexpected mutation: %v", result)
	}
}
