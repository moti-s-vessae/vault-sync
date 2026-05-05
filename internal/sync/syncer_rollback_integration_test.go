package sync_test

import (
	"log"
	"os"
	"testing"

	"github.com/your-org/vault-sync/internal/sync"
	"github.com/your-org/vault-sync/internal/vault"
)

// TestRollbackGuard_IntegrationWithSyncer verifies that a guard correctly
// preserves a pre-sync baseline and can restore it after a failed sync.
func TestRollbackGuard_IntegrationWithSyncer(t *testing.T) {
	logger := log.New(os.Stdout, "[test] ", 0)
	manager := vault.NewRollbackManager(5)
	guard := sync.NewRollbackGuard(manager, logger)

	baseline := map[string]string{
		"DB_HOST": "localhost",
		"DB_PORT": "5432",
	}

	// Simulate: save baseline before sync attempt
	guard.Before("baseline", baseline)

	// Simulate: save a partial/failed sync state
	partial := map[string]string{
		"DB_HOST": "broken-host",
	}
	guard.Before("failed-sync", partial)

	if guard.Depth() != 2 {
		t.Fatalf("expected 2 snapshots, got %d", guard.Depth())
	}

	// Restore to baseline
	restored, err := guard.Restore()
	if err != nil {
		t.Fatalf("restore failed: %v", err)
	}

	if restored["DB_HOST"] != "localhost" {
		t.Errorf("expected DB_HOST=localhost after rollback, got %q", restored["DB_HOST"])
	}
	if restored["DB_PORT"] != "5432" {
		t.Errorf("expected DB_PORT=5432 after rollback, got %q", restored["DB_PORT"])
	}
	if guard.Depth() != 1 {
		t.Errorf("expected depth 1 after rollback, got %d", guard.Depth())
	}
}
