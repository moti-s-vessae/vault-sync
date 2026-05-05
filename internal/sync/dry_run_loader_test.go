package sync_test

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/your-org/vault-sync/internal/sync"
)

type stubLoader struct {
	secrets map[string]string
	err     error
}

func (s *stubLoader) Load(_ context.Context, _ string) (map[string]string, error) {
	return s.secrets, s.err
}

func TestNewDryRunLoader_NilInner(t *testing.T) {
	_, err := sync.NewDryRunLoader(nil, nil)
	if err == nil {
		t.Fatal("expected error for nil inner loader")
	}
}

func TestDryRunLoader_Load_LogsCount(t *testing.T) {
	var buf bytes.Buffer
	inner := &stubLoader{secrets: map[string]string{"A": "1", "B": "2"}}
	l, err := sync.NewDryRunLoader(inner, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := l.Load(context.Background(), "secret/app")
	if err != nil {
		t.Fatalf("unexpected load error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 secrets, got %d", len(got))
	}

	out := buf.String()
	if !strings.Contains(out, "2 secret(s)") {
		t.Errorf("expected count in output, got: %s", out)
	}
	if !strings.Contains(out, "secret/app") {
		t.Errorf("expected path in output, got: %s", out)
	}
}

func TestDryRunLoader_Load_PropagatesError(t *testing.T) {
	var buf bytes.Buffer
	sentinel := errors.New("vault unavailable")
	inner := &stubLoader{err: sentinel}
	l, _ := sync.NewDryRunLoader(inner, &buf)

	_, err := l.Load(context.Background(), "secret/app")
	if !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel error, got: %v", err)
	}
	if !strings.Contains(buf.String(), "error") {
		t.Errorf("expected error in log output, got: %s", buf.String())
	}
}

func TestDryRunLoader_DefaultsToStdout(t *testing.T) {
	inner := &stubLoader{secrets: map[string]string{}}
	l, err := sync.NewDryRunLoader(inner, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should not panic.
	_, _ = l.Load(context.Background(), "secret/app")
}
