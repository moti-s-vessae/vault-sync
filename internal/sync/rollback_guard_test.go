package sync_test

import (
	"log"
	"os"
	"testing"

	"github.com/your-org/vault-sync/internal/sync"
	"github.com/your-org/vault-sync/internal/vault"
)

func newTestGuard(depth int) *sync.RollbackGuard {
	m := vault.NewRollbackManager(depth)
	l := log.New(os.Stdout, "", 0)
	return sync.NewRollbackGuard(m, l)
}

func TestRollbackGuard_BeforeAndRestore(t *testing.T) {
	g := newTestGuard(5)
	old := map[string]string{"FOO": "bar"}
	g.Before("pre-sync", old)
	g.Before("post-sync", map[string]string{"FOO": "new"})

	restored, err := g.Restore()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if restored["FOO"] != "bar" {
		t.Errorf("expected restored FOO=bar, got %q", restored["FOO"])
	}
}

func TestRollbackGuard_Restore_NothingToRollback(t *testing.T) {
	g := newTestGuard(5)
	g.Before("only", map[string]string{})

	_, err := g.Restore()
	if err == nil {
		t.Error("expected error when only one snapshot exists")
	}
}

func TestRollbackGuard_Depth(t *testing.T) {
	g := newTestGuard(5)
	if g.Depth() != 0 {
		t.Errorf("expected depth 0, got %d", g.Depth())
	}
	g.Before("v1", map[string]string{})
	if g.Depth() != 1 {
		t.Errorf("expected depth 1, got %d", g.Depth())
	}
}

func TestRollbackGuard_SnapshotIsolation(t *testing.T) {
	g := newTestGuard(5)
	secrets := map[string]string{"KEY": "original"}
	g.Before("v1", secrets)
	secrets["KEY"] = "mutated"
	g.Before("v2", secrets)

	restored, err := g.Restore()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if restored["KEY"] != "original" {
		t.Errorf("expected original, got %q", restored["KEY"])
	}
}
