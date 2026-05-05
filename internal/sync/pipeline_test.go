package sync

import (
	"context"
	"errors"
	"testing"

	"github.com/your-org/vault-sync/internal/vault"
)

type stubLoader struct {
	secrets map[string]string
	err     error
}

func (s *stubLoader) Load(_ context.Context, _ string) (map[string]string, error) {
	return s.secrets, s.err
}

func TestNewPipeline_NilLoader(t *testing.T) {
	_, err := NewPipeline(nil)
	if err == nil {
		t.Fatal("expected error for nil loader")
	}
}

func TestPipeline_Run_NoStages(t *testing.T) {
	loader := &stubLoader{secrets: map[string]string{"KEY": "val"}}
	p, err := NewPipeline(loader)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, err := p.Run(context.Background(), "secret/app")
	if err != nil {
		t.Fatalf("unexpected run error: %v", err)
	}
	if got["KEY"] != "val" {
		t.Errorf("expected KEY=val, got %v", got)
	}
}

func TestPipeline_Run_LoaderError(t *testing.T) {
	loader := &stubLoader{err: errors.New("vault down")}
	p, _ := NewPipeline(loader)
	_, err := p.Run(context.Background(), "secret/app")
	if err == nil || !errors.Is(err, err) {
		t.Fatal("expected propagated loader error")
	}
}

func TestPipeline_Run_StageError(t *testing.T) {
	loader := &stubLoader{secrets: map[string]string{"K": "v"}}
	badStage := PipelineStage{
		Name: "fail",
		Apply: func(_ map[string]string) (map[string]string, error) {
			return nil, errors.New("stage failed")
		},
	}
	p, _ := NewPipeline(loader, badStage)
	_, err := p.Run(context.Background(), "secret/app")
	if err == nil {
		t.Fatal("expected stage error")
	}
}

func TestPipeline_Run_FilterStage(t *testing.T) {
	loader := &stubLoader{secrets: map[string]string{
		"APP_KEY": "1",
		"DB_PASS": "2",
	}}
	p, _ := NewPipeline(loader, FilterStage([]string{"APP_"}))
	got, err := p.Run(context.Background(), "secret/app")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := got["DB_PASS"]; ok {
		t.Error("expected DB_PASS to be filtered out")
	}
	if got["APP_KEY"] != "1" {
		t.Errorf("expected APP_KEY=1, got %v", got)
	}
}

func TestPipeline_Run_RenameStage(t *testing.T) {
	loader := &stubLoader{secrets: map[string]string{"OLD_KEY": "value"}}
	rules := []vault.RenameRule{{From: "OLD_KEY", To: "NEW_KEY"}}
	p, _ := NewPipeline(loader, RenameStage(rules))
	got, err := p.Run(context.Background(), "secret/app")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got["NEW_KEY"] != "value" {
		t.Errorf("expected NEW_KEY=value, got %v", got)
	}
	if _, ok := got["OLD_KEY"]; ok {
		t.Error("OLD_KEY should have been renamed")
	}
}

func TestPipeline_Run_MultipleStages(t *testing.T) {
	loader := &stubLoader{secrets: map[string]string{
		"APP_FOO": "bar",
		"APP_BAZ": "qux",
		"SKIP":    "me",
	}}
	rules := []vault.RenameRule{{From: "APP_FOO", To: "FOO"}}
	p, _ := NewPipeline(loader,
		FilterStage([]string{"APP_"}),
		RenameStage(rules),
	)
	got, err := p.Run(context.Background(), "secret/app")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := got["SKIP"]; ok {
		t.Error("SKIP should have been filtered")
	}
	if got["FOO"] != "bar" {
		t.Errorf("expected FOO=bar after rename, got %v", got)
	}
}
