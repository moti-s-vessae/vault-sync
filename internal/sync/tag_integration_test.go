package sync

import (
	"testing"
)

func TestTagStage_Integration_WithPipeline(t *testing.T) {
	loader := &mockLoader{
		secrets: map[string]string{
			"DB_HOST":  "localhost",
			"DB_PASS":  "secret",
			"APP_PORT": "8080",
		},
	}

	tagger, err := NewSecretTagger([]TagRule{
		{Pattern: "^DB_", Tag: "db"},
		{Pattern: "^APP_", Tag: "app"},
	})
	if err != nil {
		t.Fatalf("NewSecretTagger: %v", err)
	}

	pipeline, err := NewPipeline(loader, TagStage(tagger))
	if err != nil {
		t.Fatalf("NewPipeline: %v", err)
	}

	result, err := pipeline.Run(t.Context())
	if err != nil {
		t.Fatalf("pipeline.Run: %v", err)
	}

	expected := map[string]string{
		"db:DB_HOST":  "localhost",
		"db:DB_PASS":  "secret",
		"app:APP_PORT": "8080",
	}
	if len(result) != len(expected) {
		t.Fatalf("expected %d keys, got %d: %v", len(expected), len(result), result)
	}
	for k, v := range expected {
		if got, ok := result[k]; !ok || got != v {
			t.Errorf("result[%q] = %q, want %q", k, got, v)
		}
	}
}

func TestTagStage_Integration_UnmatchedKeysPassThrough(t *testing.T) {
	loader := &mockLoader{
		secrets: map[string]string{
			"UNRELATED_KEY": "value",
			"DB_HOST":       "localhost",
		},
	}

	tagger, _ := NewSecretTagger([]TagRule{{Pattern: "^DB_", Tag: "db"}})

	pipeline, err := NewPipeline(loader, TagStage(tagger))
	if err != nil {
		t.Fatalf("NewPipeline: %v", err)
	}

	result, err := pipeline.Run(t.Context())
	if err != nil {
		t.Fatalf("pipeline.Run: %v", err)
	}

	if _, ok := result["db:DB_HOST"]; !ok {
		t.Error("expected db:DB_HOST in result")
	}
	if _, ok := result["UNRELATED_KEY"]; !ok {
		t.Error("expected UNRELATED_KEY unchanged in result")
	}
}
