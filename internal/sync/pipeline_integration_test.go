package sync_test

import (
	"context"
	"testing"

	"github.com/your-org/vault-sync/internal/sync"
	"github.com/your-org/vault-sync/internal/vault"
)

type integStubLoader struct {
	secrets map[string]string
}

func (l *integStubLoader) Load(_ context.Context, _ string) (map[string]string, error) {
	out := make(map[string]string, len(l.secrets))
	for k, v := range l.secrets {
		out[k] = v
	}
	return out, nil
}

func TestPipeline_Integration_FilterRenameTransform(t *testing.T) {
	loader := &integStubLoader{
		secrets: map[string]string{
			"APP_DATABASE_URL": "postgres://localhost",
			"APP_SECRET_KEY":  "s3cr3t",
			"IGNORE_ME":       "noise",
		},
	}

	renameRules := []vault.RenameRule{
		{From: "APP_DATABASE_URL", To: "DATABASE_URL"},
	}
	transformRules := []vault.TransformRule{
		{Pattern: "^APP_", Replacement: ""},
	}

	p, err := sync.NewPipeline(
		loader,
		sync.FilterStage([]string{"APP_"}),
		sync.RenameStage(renameRules),
		sync.TransformStage(transformRules),
	)
	if err != nil {
		t.Fatalf("NewPipeline: %v", err)
	}

	got, err := p.Run(context.Background(), "secret/app")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if _, ok := got["IGNORE_ME"]; ok {
		t.Error("IGNORE_ME should have been filtered")
	}
	if got["DATABASE_URL"] != "postgres://localhost" {
		t.Errorf("expected DATABASE_URL=postgres://localhost, got %v", got["DATABASE_URL"])
	}
	// APP_SECRET_KEY -> after rename (no rule) stays APP_SECRET_KEY -> after transform becomes SECRET_KEY
	if got["SECRET_KEY"] != "s3cr3t" {
		t.Errorf("expected SECRET_KEY=s3cr3t, got %v", got)
	}
}
