package sync_test

import (
	"errors"
	"testing"

	"github.com/your-org/vault-sync/internal/sync"
	"github.com/your-org/vault-sync/internal/vault"
)

type mockPolicyChecker struct {
	allowedPaths map[string]bool
	err          error
}

func (m *mockPolicyChecker) CheckAccess(path string) error {
	if m.err != nil {
		return m.err
	}
	if !m.allowedPaths[path] {
		return vault.ErrAccessDenied
	}
	return nil
}

func TestPolicyGuard_AllPathsAllowed(t *testing.T) {
	checker := &mockPolicyChecker{
		allowedPaths: map[string]bool{
			"secret/app/db": true,
			"secret/app/api": true,
		},
	}
	guard := sync.NewPolicyGuard(checker)
	paths := []string{"secret/app/db", "secret/app/api"}
	if err := guard.EnforceAll(paths); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestPolicyGuard_OnePathDenied(t *testing.T) {
	checker := &mockPolicyChecker{
		allowedPaths: map[string]bool{
			"secret/app/db": true,
		},
	}
	guard := sync.NewPolicyGuard(checker)
	paths := []string{"secret/app/db", "secret/app/restricted"}
	if err := guard.EnforceAll(paths); err == nil {
		t.Fatal("expected error for denied path, got nil")
	}
}

func TestPolicyGuard_CheckerError(t *testing.T) {
	checker := &mockPolicyChecker{
		err: errors.New("vault unreachable"),
	}
	guard := sync.NewPolicyGuard(checker)
	paths := []string{"secret/app/db"}
	if err := guard.EnforceAll(paths); err == nil {
		t.Fatal("expected error from checker, got nil")
	}
}

func TestPolicyGuard_EmptyPaths(t *testing.T) {
	checker := &mockPolicyChecker{}
	guard := sync.NewPolicyGuard(checker)
	if err := guard.EnforceAll([]string{}); err != nil {
		t.Fatalf("expected no error for empty paths, got: %v", err)
	}
}
