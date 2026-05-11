package sync_test

import (
	"context"
	"testing"

	"github.com/your-org/vault-sync/internal/sync"
)

type staticNormLoader struct {
	data map[string]string
}

func (l *staticNormLoader) Load(_ context.Context, _ string) (map[string]string, error) {
	out := make(map[string]string, len(l.data))
	for k, v := range l.data {
		out[k] = v
	}
	return out, nil
}

func TestNormalizeStage_Integration_WithPipeline(t *testing.T) {
	loader := &staticNormLoader{
		data: map[string]string{
			"db-host": "localhost",
			"api-key": "secret",
		},
	}

	normalizer, err := sync.NewSecretNormalizer([]sync.NormalizeRule{
		{Pattern: ".*", Strategy: "snake"},
	})
	if err != nil {
		t.Fatalf("failed to create normalizer: %v", err)
	}

	pipeline, err := sync.NewPipeline(loader, sync.NormalizeStage(normalizer))
	if err != nil {
		t.Fatalf("failed to create pipeline: %v", err)
	}

	result, err := pipeline.Run(context.Background(), "secret/app")
	if err != nil {
		t.Fatalf("pipeline run failed: %v", err)
	}

	if result["DB_HOST"] != "localhost" {
		t.Errorf("expected DB_HOST=localhost, got %q", result["DB_HOST"])
	}
	if result["API_KEY"] != "secret" {
		t.Errorf("expected API_KEY=secret, got %q", result["API_KEY"])
	}
	if _, ok := result["db-host"]; ok {
		t.Error("original key db-host should not be present after normalization")
	}
}

func TestNormalizeStage_Integration_CollisionPropagatesError(t *testing.T) {
	loader := &staticNormLoader{
		data: map[string]string{
			"db_host": "a",
			"DB_HOST": "b",
		},
	}

	normalizer, err := sync.NewSecretNormalizer([]sync.NormalizeRule{
		{Pattern: ".*", Strategy: "upper"},
	})
	if err != nil {
		t.Fatalf("failed to create normalizer: %v", err)
	}

	pipeline, err := sync.NewPipeline(loader, sync.NormalizeStage(normalizer))
	if err != nil {
		t.Fatalf("failed to create pipeline: %v", err)
	}

	_, err = pipeline.Run(context.Background(), "secret/app")
	if err == nil {
		t.Fatal("expected collision error from pipeline")
	}
}
