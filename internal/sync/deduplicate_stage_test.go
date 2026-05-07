package sync

import (
	"testing"
)

func TestDeduplicateStage_Integration_WithPipeline(t *testing.T) {
	// Simulate a pipeline stage that validates and passes secrets through.
	stage := DeduplicateStage(DeduplicateKeepLast)

	input := map[string]string{
		"SERVICE_URL": "https://api.example.com",
		"TIMEOUT":     "30s",
	}

	out, err := stage(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out) != len(input) {
		t.Errorf("expected %d keys, got %d", len(input), len(out))
	}
	for k, v := range input {
		if out[k] != v {
			t.Errorf("key %q: expected %q, got %q", k, v, out[k])
		}
	}
}

func TestDeduplicateStage_Integration_EmptySecrets(t *testing.T) {
	stage := DeduplicateStage(DeduplicateKeepFirst)
	out, err := stage(map[string]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out) != 0 {
		t.Errorf("expected empty output, got %v", out)
	}
}

func TestDeduplicateStage_AllStrategies_SingleMap(t *testing.T) {
	strategies := []DeduplicateStrategy{
		DeduplicateKeepFirst,
		DeduplicateKeepLast,
		DeduplicateError,
	}
	input := map[string]string{"ONLY_KEY": "value"}

	for _, s := range strategies {
		t.Run(string(s), func(t *testing.T) {
			stage := DeduplicateStage(s)
			out, err := stage(input)
			if err != nil {
				t.Fatalf("strategy %q: unexpected error: %v", s, err)
			}
			if out["ONLY_KEY"] != "value" {
				t.Errorf("strategy %q: expected value, got %q", s, out["ONLY_KEY"])
			}
		})
	}
}
