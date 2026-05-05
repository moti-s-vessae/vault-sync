package sync

import (
	"fmt"
	"log"

	"github.com/your-org/vault-sync/internal/vault"
)

// RollbackGuard wraps a RollbackManager and provides pre/post-sync hooks
// that automatically snapshot secrets before a sync and restore on failure.
type RollbackGuard struct {
	manager *vault.RollbackManager
	logger  *log.Logger
}

// NewRollbackGuard creates a RollbackGuard backed by the given manager and logger.
func NewRollbackGuard(manager *vault.RollbackManager, logger *log.Logger) *RollbackGuard {
	return &RollbackGuard{manager: manager, logger: logger}
}

// Before saves a snapshot labelled with the provided tag before a sync begins.
func (g *RollbackGuard) Before(tag string, current map[string]string) {
	g.manager.Save(tag, current)
	g.logger.Printf("[rollback] snapshot saved: %s (%d keys)", tag, len(current))
}

// Restore rolls back to the previous snapshot and returns the restored secrets.
// Returns an error if rollback is not possible.
func (g *RollbackGuard) Restore() (map[string]string, error) {
	snap, err := g.manager.Rollback()
	if err != nil {
		return nil, fmt.Errorf("rollback failed: %w", err)
	}
	g.logger.Printf("[rollback] restored snapshot: %s (created %s)", snap.Label, snap.CreatedAt.Format("15:04:05"))
	return snap.Secrets, nil
}

// Depth returns how many snapshots are currently held.
func (g *RollbackGuard) Depth() int {
	return g.manager.Depth()
}
