package vault

import (
	"testing"
)

func TestDiffSecrets_Added(t *testing.T) {
	old := map[string]string{}
	new_ := map[string]string{"FOO": "bar"}
	changes := DiffSecrets(old, new_)
	if len(changes) != 1 || changes[0].Action != "added" || changes[0].Key != "FOO" {
		t.Errorf("expected one added change, got %+v", changes)
	}
}

func TestDiffSecrets_Removed(t *testing.T) {
	old := map[string]string{"FOO": "bar"}
	new_ := map[string]string{}
	changes := DiffSecrets(old, new_)
	if len(changes) != 1 || changes[0].Action != "removed" || changes[0].Key != "FOO" {
		t.Errorf("expected one removed change, got %+v", changes)
	}
}

func TestDiffSecrets_Changed(t *testing.T) {
	old := map[string]string{"FOO": "old"}
	new_ := map[string]string{"FOO": "new"}
	changes := DiffSecrets(old, new_)
	if len(changes) != 1 || changes[0].Action != "changed" {
		t.Errorf("expected one changed entry, got %+v", changes)
	}
	if changes[0].OldVal != "old" || changes[0].NewVal != "new" {
		t.Errorf("unexpected old/new values: %+v", changes[0])
	}
}

func TestDiffSecrets_Unchanged(t *testing.T) {
	old := map[string]string{"FOO": "same"}
	new_ := map[string]string{"FOO": "same"}
	changes := DiffSecrets(old, new_)
	if len(changes) != 1 || changes[0].Action != "unchanged" {
		t.Errorf("expected unchanged, got %+v", changes)
	}
}

func TestDiffSecrets_Empty(t *testing.T) {
	changes := DiffSecrets(map[string]string{}, map[string]string{})
	if len(changes) != 0 {
		t.Errorf("expected no changes, got %+v", changes)
	}
}

func TestDiffSecrets_Mixed(t *testing.T) {
	old := map[string]string{"A": "1", "B": "2", "C": "3"}
	new_ := map[string]string{"A": "1", "B": "updated", "D": "4"}
	changes := DiffSecrets(old, new_)

	actions := map[string]string{}
	for _, c := range changes {
		actions[c.Key] = c.Action
	}

	if actions["A"] != "unchanged" {
		t.Errorf("A should be unchanged, got %s", actions["A"])
	}
	if actions["B"] != "changed" {
		t.Errorf("B should be changed, got %s", actions["B"])
	}
	if actions["C"] != "removed" {
		t.Errorf("C should be removed, got %s", actions["C"])
	}
	if actions["D"] != "added" {
		t.Errorf("D should be added, got %s", actions["D"])
	}
}

func TestHasChanges_True(t *testing.T) {
	changes := []SecretChange{{Key: "X", Action: "added"}}
	if !HasChanges(changes) {
		t.Error("expected HasChanges to return true")
	}
}

func TestHasChanges_False(t *testing.T) {
	changes := []SecretChange{{Key: "X", Action: "unchanged"}}
	if HasChanges(changes) {
		t.Error("expected HasChanges to return false")
	}
}
