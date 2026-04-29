package vault

import (
	"testing"
)

func TestDiffSecrets_Added(t *testing.T) {
	current := map[string]string{"FOO": "bar", "NEW_KEY": "new"}
	existing := map[string]string{"FOO": "bar"}

	result := DiffSecrets(current, existing)

	if len(result.Added) != 1 || result.Added["NEW_KEY"] != "new" {
		t.Errorf("expected NEW_KEY in Added, got %v", result.Added)
	}
	if len(result.Removed) != 0 {
		t.Errorf("expected no removals, got %v", result.Removed)
	}
}

func TestDiffSecrets_Removed(t *testing.T) {
	current := map[string]string{"FOO": "bar"}
	existing := map[string]string{"FOO": "bar", "OLD_KEY": "old"}

	result := DiffSecrets(current, existing)

	if len(result.Removed) != 1 || result.Removed["OLD_KEY"] != "old" {
		t.Errorf("expected OLD_KEY in Removed, got %v", result.Removed)
	}
}

func TestDiffSecrets_Changed(t *testing.T) {
	current := map[string]string{"FOO": "new_val"}
	existing := map[string]string{"FOO": "old_val"}

	result := DiffSecrets(current, existing)

	if len(result.Changed) != 1 || result.Changed["FOO"] != "new_val" {
		t.Errorf("expected FOO in Changed, got %v", result.Changed)
	}
}

func TestDiffSecrets_Unchanged(t *testing.T) {
	current := map[string]string{"FOO": "bar"}
	existing := map[string]string{"FOO": "bar"}

	result := DiffSecrets(current, existing)

	if len(result.Unchanged) != 1 {
		t.Errorf("expected FOO in Unchanged, got %v", result.Unchanged)
	}
	if result.HasChanges() {
		t.Error("expected HasChanges to be false")
	}
}

func TestDiffSecrets_Empty(t *testing.T) {
	result := DiffSecrets(map[string]string{}, map[string]string{})

	if result.HasChanges() {
		t.Error("expected no changes for empty maps")
	}
}

func TestHasChanges_True(t *testing.T) {
	current := map[string]string{"A": "1"}
	existing := map[string]string{}

	result := DiffSecrets(current, existing)
	if !result.HasChanges() {
		t.Error("expected HasChanges to be true")
	}
}
