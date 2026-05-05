package vault

import (
	"fmt"
	"time"
)

// Snapshot holds a point-in-time copy of secrets for rollback purposes.
type Snapshot struct {
	Secrets   map[string]string
	CreatedAt time.Time
	Label     string
}

// RollbackManager stores snapshots and allows restoring previous secret states.
type RollbackManager struct {
	snapshots []Snapshot
	maxDepth  int
}

// NewRollbackManager creates a RollbackManager that retains up to maxDepth snapshots.
func NewRollbackManager(maxDepth int) *RollbackManager {
	if maxDepth <= 0 {
		maxDepth = 5
	}
	return &RollbackManager{maxDepth: maxDepth}
}

// Save stores a labelled snapshot of the provided secrets map.
func (r *RollbackManager) Save(label string, secrets map[string]string) {
	copy := make(map[string]string, len(secrets))
	for k, v := range secrets {
		copy[k] = v
	}
	r.snapshots = append(r.snapshots, Snapshot{
		Secrets:   copy,
		CreatedAt: time.Now(),
		Label:     label,
	})
	if len(r.snapshots) > r.maxDepth {
		r.snapshots = r.snapshots[len(r.snapshots)-r.maxDepth:]
	}
}

// Latest returns the most recently saved snapshot, or an error if none exist.
func (r *RollbackManager) Latest() (Snapshot, error) {
	if len(r.snapshots) == 0 {
		return Snapshot{}, fmt.Errorf("no snapshots available")
	}
	return r.snapshots[len(r.snapshots)-1], nil
}

// Rollback removes the latest snapshot and returns the one before it.
// Returns an error if fewer than two snapshots exist.
func (r *RollbackManager) Rollback() (Snapshot, error) {
	if len(r.snapshots) < 2 {
		return Snapshot{}, fmt.Errorf("nothing to roll back to")
	}
	r.snapshots = r.snapshots[:len(r.snapshots)-1]
	return r.snapshots[len(r.snapshots)-1], nil
}

// Depth returns the number of stored snapshots.
func (r *RollbackManager) Depth() int {
	return len(r.snapshots)
}
