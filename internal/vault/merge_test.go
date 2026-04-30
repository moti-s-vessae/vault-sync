package vault

import (
	"testing"
)

func TestMergeSecrets_Overwrite(t *testing.T) {
	base := map[string]string{"FOO": "old", "BAR": "keep"}
	incoming := map[string]string{"FOO": "new", "BAZ": "added"}

	res, err := MergeSecrets(base, incoming, MergeStrategyOverwrite)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Secrets["FOO"] != "new" {
		t.Errorf("expected FOO=new, got %s", res.Secrets["FOO"])
	}
	if res.Secrets["BAR"] != "keep" {
		t.Errorf("expected BAR=keep, got %s", res.Secrets["BAR"])
	}
	if res.Secrets["BAZ"] != "added" {
		t.Errorf("expected BAZ=added, got %s", res.Secrets["BAZ"])
	}
	if len(res.Conflicts) != 1 || res.Conflicts[0].Key != "FOO" {
		t.Errorf("expected one conflict on FOO, got %+v", res.Conflicts)
	}
}

func TestMergeSecrets_KeepExisting(t *testing.T) {
	base := map[string]string{"FOO": "local"}
	incoming := map[string]string{"FOO": "remote"}

	res, err := MergeSecrets(base, incoming, MergeStrategyKeepExisting)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Secrets["FOO"] != "local" {
		t.Errorf("expected FOO=local, got %s", res.Secrets["FOO"])
	}
	if len(res.Conflicts) != 1 {
		t.Errorf("expected 1 conflict, got %d", len(res.Conflicts))
	}
}

func TestMergeSecrets_Error_OnConflict(t *testing.T) {
	base := map[string]string{"SECRET": "a"}
	incoming := map[string]string{"SECRET": "b"}

	_, err := MergeSecrets(base, incoming, MergeStrategyError)
	if err == nil {
		t.Fatal("expected error on conflict, got nil")
	}
	mergeErr, ok := err.(*MergeError)
	if !ok {
		t.Fatalf("expected *MergeError, got %T", err)
	}
	if len(mergeErr.Conflicts) != 1 || mergeErr.Conflicts[0].Key != "SECRET" {
		t.Errorf("unexpected conflicts: %+v", mergeErr.Conflicts)
	}
}

func TestMergeSecrets_NoConflict(t *testing.T) {
	base := map[string]string{"A": "1"}
	incoming := map[string]string{"A": "1", "B": "2"}

	res, err := MergeSecrets(base, incoming, MergeStrategyError)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Conflicts) != 0 {
		t.Errorf("expected no conflicts, got %+v", res.Conflicts)
	}
	if res.Secrets["B"] != "2" {
		t.Errorf("expected B=2, got %s", res.Secrets["B"])
	}
}

func TestMergeSecrets_EmptyBase(t *testing.T) {
	base := map[string]string{}
	incoming := map[string]string{"X": "1", "Y": "2"}

	res, err := MergeSecrets(base, incoming, MergeStrategyOverwrite)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Secrets) != 2 {
		t.Errorf("expected 2 secrets, got %d", len(res.Secrets))
	}
	if len(res.Conflicts) != 0 {
		t.Errorf("expected no conflicts, got %+v", res.Conflicts)
	}
}
