package sync

import (
	"testing"
)

func TestValidateDeduplicateStrategy_Valid(t *testing.T) {
	for _, s := range []DeduplicateStrategy{
		DeduplicateKeepFirst,
		DeduplicateKeepLast,
		DeduplicateError,
	} {
		if err := ValidateDeduplicateStrategy(s); err != nil {
			t.Errorf("expected no error for %q, got %v", s, err)
		}
	}
}

func TestValidateDeduplicateStrategy_Invalid(t *testing.T) {
	if err := ValidateDeduplicateStrategy("bogus"); err == nil {
		t.Error("expected error for unknown strategy")
	}
}

func TestDeduplicateSecrets_KeepFirst(t *testing.T) {
	a := map[string]string{"KEY": "first", "A": "1"}
	b := map[string]string{"KEY": "second", "B": "2"}

	got, err := DeduplicateSecrets([]map[string]string{a, b}, DeduplicateKeepFirst)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got["KEY"] != "first" {
		t.Errorf("expected first, got %q", got["KEY"])
	}
	if got["A"] != "1" || got["B"] != "2" {
		t.Error("expected non-duplicate keys to be present")
	}
}

func TestDeduplicateSecrets_KeepLast(t *testing.T) {
	a := map[string]string{"KEY": "first"}
	b := map[string]string{"KEY": "second"}

	got, err := DeduplicateSecrets([]map[string]string{a, b}, DeduplicateKeepLast)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got["KEY"] != "second" {
		t.Errorf("expected second, got %q", got["KEY"])
	}
}

func TestDeduplicateSecrets_Error_OnConflict(t *testing.T) {
	a := map[string]string{"KEY": "v1"}
	b := map[string]string{"KEY": "v2"}

	_, err := DeduplicateSecrets([]map[string]string{a, b}, DeduplicateError)
	if err == nil {
		t.Error("expected error on duplicate key")
	}
}

func TestDeduplicateSecrets_NoConflict(t *testing.T) {
	a := map[string]string{"X": "1"}
	b := map[string]string{"Y": "2"}

	got, err := DeduplicateSecrets([]map[string]string{a, b}, DeduplicateError)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 keys, got %d", len(got))
	}
}

func TestDeduplicateSecrets_EmptyInput(t *testing.T) {
	got, err := DeduplicateSecrets(nil, DeduplicateKeepFirst)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty map, got %v", got)
	}
}

func TestDeduplicateSecrets_InvalidStrategy(t *testing.T) {
	_, err := DeduplicateSecrets([]map[string]string{{"K": "v"}}, "invalid")
	if err == nil {
		t.Error("expected error for invalid strategy")
	}
}

func TestDeduplicateStage_PassThrough(t *testing.T) {
	stage := DeduplicateStage(DeduplicateKeepFirst)
	input := map[string]string{"FOO": "bar"}

	out, err := stage(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out["FOO"] != "bar" {
		t.Errorf("expected bar, got %q", out["FOO"])
	}
}

func TestDeduplicateStage_InvalidStrategy_ReturnsError(t *testing.T) {
	stage := DeduplicateStage("nope")
	_, err := stage(map[string]string{})
	if err == nil {
		t.Error("expected error for invalid strategy")
	}
}
