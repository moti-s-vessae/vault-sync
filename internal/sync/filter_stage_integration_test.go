package sync_test

import (
	"testing"

	"github.com/your-org/vault-sync/internal/sync"
)

type staticLoader struct {
	secrets map[string]string
}

func (s *staticLoader) Load(path string) (map[string]string, error) {
	out := make(map[string]string, len(s.secrets))
	for k, v := range s.secrets {
		out[k] = v
	}
	return out, nil
}

func TestKeyFilterStage_Integration_WithPipeline(t *testing.T) {
	loader := &staticLoader{
		secrets: map[string]string{
			"DB_HOST":        "localhost",
			"DB_PASS":        "s3cr3t",
			"REDIS_URL":      "redis://localhost",
			"INTERNAL_TOKEN": "tok",
		},
	}

	rules := []sync.FilterRule{
		{Pattern: "^DB_"},
	}

	pipeline, err := sync.NewPipeline(loader, sync.KeyFilterStage(rules))
	if err != nil {
		t.Fatalf("unexpected error building pipeline: %v", err)
	}

	result, err := pipeline.Run("secret/app")
	if err != nil {
		t.Fatalf("unexpected error running pipeline: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("expected 2 secrets, got %d", len(result))
	}
	for k := range result {
		if k != "DB_HOST" && k != "DB_PASS" {
			t.Errorf("unexpected key in result: %s", k)
		}
	}
}

func TestKeyFilterStage_Integration_NegateWithPipeline(t *testing.T) {
	loader := &staticLoader{
		secrets: map[string]string{
			"PUBLIC_URL":     "https://example.com",
			"INTERNAL_TOKEN": "tok",
			"INTERNAL_KEY":   "key",
		},
	}

	rules := []sync.FilterRule{
		{Pattern: "^INTERNAL_", Negate: true},
	}

	pipeline, err := sync.NewPipeline(loader, sync.KeyFilterStage(rules))
	if err != nil {
		t.Fatalf("unexpected error building pipeline: %v", err)
	}

	result, err := pipeline.Run("secret/app")
	if err != nil {
		t.Fatalf("unexpected error running pipeline: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 secret, got %d", len(result))
	}
	if _, ok := result["PUBLIC_URL"]; !ok {
		t.Error("expected PUBLIC_URL in result")
	}
}
