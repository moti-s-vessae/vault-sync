package sync

import (
	"strings"
	"testing"
)

func TestNewSecretTruncator_NoRules(t *testing.T) {
	_, err := NewSecretTruncator(nil)
	if err == nil {
		t.Fatal("expected error for empty rules")
	}
}

func TestNewSecretTruncator_EmptyPattern(t *testing.T) {
	_, err := NewSecretTruncator([]TruncateRule{{Pattern: "", MaxLen: 10}})
	if err == nil {
		t.Fatal("expected error for empty pattern")
	}
}

func TestNewSecretTruncator_NonPositiveMaxLen(t *testing.T) {
	_, err := NewSecretTruncator([]TruncateRule{{Pattern: "*", MaxLen: 0}})
	if err == nil {
		t.Fatal("expected error for zero max_len")
	}
}

func TestNewSecretTruncator_Valid(t *testing.T) {
	tr, err := NewSecretTruncator([]TruncateRule{{Pattern: "*", MaxLen: 32}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tr == nil {
		t.Fatal("expected non-nil truncator")
	}
}

func TestSecretTruncator_Apply_TruncatesMatchingKey(t *testing.T) {
	tr, _ := NewSecretTruncator([]TruncateRule{
		{Pattern: "*_TOKEN", MaxLen: 5, Suffix: "..."},
	})
	secrets := map[string]string{
		"API_TOKEN": "abcdefghij",
		"DB_PASS":   "short",
	}
	out := tr.Apply(secrets)
	if got := out["API_TOKEN"]; got != "abcde..." {
		t.Errorf("expected \"abcde...\", got %q", got)
	}
	if got := out["DB_PASS"]; got != "short" {
		t.Errorf("expected \"short\" unchanged, got %q", got)
	}
}

func TestSecretTruncator_Apply_NoTruncationWhenUnderLimit(t *testing.T) {
	tr, _ := NewSecretTruncator([]TruncateRule{
		{Pattern: "*", MaxLen: 100, Suffix: "..."},
	})
	secrets := map[string]string{"KEY": "hello"}
	out := tr.Apply(secrets)
	if out["KEY"] != "hello" {
		t.Errorf("expected value unchanged, got %q", out["KEY"])
	}
}

func TestSecretTruncator_Apply_DoesNotMutateInput(t *testing.T) {
	tr, _ := NewSecretTruncator([]TruncateRule{
		{Pattern: "*", MaxLen: 3},
	})
	orig := map[string]string{"KEY": "abcdefgh"}
	_ = tr.Apply(orig)
	if orig["KEY"] != "abcdefgh" {
		t.Error("original map was mutated")
	}
}

func TestSecretTruncator_Apply_FirstMatchWins(t *testing.T) {
	tr, _ := NewSecretTruncator([]TruncateRule{
		{Pattern: "DB_*", MaxLen: 4, Suffix: "-"},
		{Pattern: "*", MaxLen: 2, Suffix: "!"},
	})
	secrets := map[string]string{"DB_PASS": "password"}
	out := tr.Apply(secrets)
	if got := out["DB_PASS"]; got != "pass-" {
		t.Errorf("expected \"pass-\", got %q", got)
	}
}

func TestTruncateStage_Integration(t *testing.T) {
	rules := []TruncateRule{
		{Pattern: "SECRET_*", MaxLen: 6, Suffix: "…"},
	}
	stage := TruncateStage(rules)
	in := map[string]string{
		"SECRET_KEY":  strings.Repeat("x", 20),
		"NORMAL_KEY": "unchanged",
	}
	out, err := stage(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := out["SECRET_KEY"]; got != "xxxxxx…" {
		t.Errorf("expected \"xxxxxx…\", got %q", got)
	}
	if got := out["NORMAL_KEY"]; got != "unchanged" {
		t.Errorf("expected \"unchanged\", got %q", got)
	}
}
