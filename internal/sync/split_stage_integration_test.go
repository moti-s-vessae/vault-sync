package sync_test

import (
	"context"
	"testing"

	"github.com/your-org/vault-sync/internal/sync"
)

type staticSplitLoader struct {
	data map[string]string
}

func (l *staticSplitLoader) Load(_ context.Context, _ string) (map[string]string, error) {
	out := make(map[string]string, len(l.data))
	for k, v := range l.data {
		out[k] = v
	}
	return out, nil
}

func TestSplitStage_Integration_WithPipeline(t *testing.T) {
	loader := &staticSplitLoader{
		data: map[string]string{
			"ALLOWED_IPS": "10.0.0.1,10.0.0.2,10.0.0.3",
			"API_KEY":     "secret123",
		},
	}

	splitter, err := sync.NewSecretSplitter([]sync.SplitRule{
		{Pattern: "ALLOWED_IPS", Separator: ",", KeyTemplate: "{{.Key}}_{{.Index}}"},
	})
	if err != nil {
		t.Fatalf("failed to create splitter: %v", err)
	}

	pipeline, err := sync.NewPipeline(loader, sync.SplitStage(splitter))
	if err != nil {
		t.Fatalf("failed to create pipeline: %v", err)
	}

	result, err := pipeline.Run(context.Background(), "secret/app")
	if err != nil {
		t.Fatalf("pipeline run failed: %v", err)
	}

	expected := map[string]string{
		"ALLOWED_IPS_1": "10.0.0.1",
		"ALLOWED_IPS_2": "10.0.0.2",
		"ALLOWED_IPS_3": "10.0.0.3",
		"API_KEY":       "secret123",
	}
	for k, v := range expected {
		if result[k] != v {
			t.Errorf("result[%q] = %q, want %q", k, result[k], v)
		}
	}
	if _, ok := result["ALLOWED_IPS"]; ok {
		t.Error("original key should have been removed after split")
	}
}
