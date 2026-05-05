package vault_test

import (
	"testing"

	"github.com/your-org/vault-sync/internal/vault"
)

func TestRollbackManager_SaveAndLatest(t *testing.T) {
	rm := vault.NewRollbackManager(5)
	secrets := map[string]string{"KEY": "value1"}
	rm.Save("first", secrets)

	snap, err := rm.Latest()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if snap.Label != "first" {
		t.Errorf("expected label 'first', got %q", snap.Label)
	}
	if snap.Secrets["KEY"] != "value1" {
		t.Errorf("expected KEY=value1, got %q", snap.Secrets["KEY"])
	}
}

func TestRollbackManager_SnapshotIsIsolated(t *testing.T) {
	rm := vault.NewRollbackManager(5)
	secrets := map[string]string{"KEY": "original"}
	rm.Save("v1", secrets)
	secrets["KEY"] = "mutated"

	snap, _ := rm.Latest()
	if snap.Secrets["KEY"] != "original" {
		t.Errorf("snapshot should be isolated from mutation, got %q", snap.Secrets["KEY"])
	}
}

func TestRollbackManager_Rollback(t *testing.T) {
	rm := vault.NewRollbackManager(5)
	rm.Save("v1", map[string]string{"A": "1"})
	rm.Save("v2", map[string]string{"A": "2"})

	snap, err := rm.Rollback()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if snap.Label != "v1" {
		t.Errorf("expected rollback to v1, got %q", snap.Label)
	}
	if rm.Depth() != 1 {
		t.Errorf("expected depth 1 after rollback, got %d", rm.Depth())
	}
}

func TestRollbackManager_Rollback_NothingToRollback(t *testing.T) {
	rm := vault.NewRollbackManager(5)
	rm.Save("only", map[string]string{})

	_, err := rm.Rollback()
	if err == nil {
		t.Error("expected error when rolling back with only one snapshot")
	}
}

func TestRollbackManager_Latest_Empty(t *testing.T) {
	rm := vault.NewRollbackManager(5)
	_, err := rm.Latest()
	if err == nil {
		t.Error("expected error when no snapshots exist")
	}
}

func TestRollbackManager_MaxDepth(t *testing.T) {
	rm := vault.NewRollbackManager(3)
	for i := 0; i < 6; i++ {
		rm.Save("snap", map[string]string{})
	}
	if rm.Depth() != 3 {
		t.Errorf("expected depth 3, got %d", rm.Depth())
	}
}
