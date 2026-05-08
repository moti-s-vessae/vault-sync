package sync

import (
	"testing"
)

func TestNewSecretRedactor_NoRules(t *testing.T) {
	_, err := NewSecretRedactor(nil)
	if err == nil {
		t.Fatal("expected error for empty rules")
	}
}

func TestNewSecretRedactor_EmptyPattern(t *testing.T) {
	_, err := NewSecretRedactor([]RedactRule{{Pattern: "", Replacement: "[REDACTED]"}})
	if err == nil {
		t.Fatal("expected error for empty pattern")
	}
}

func TestNewSecretRedactor_EmptyReplacement(t *testing.T) {
	_, err := NewSecretRedactor([]RedactRule{{Pattern: ".*", Replacement: ""}})
	if err == nil {
		t.Fatal("expected error for empty replacement")
	}
}

func TestNewSecretRedactor_InvalidPattern(t *testing.T) {
	_, err := NewSecretRedactor([]RedactRule{{Pattern: "[", Replacement: "[REDACTED]"}})
	if err == nil {
		t.Fatal("expected error for invalid regex")
	}
}

func TestSecretRedactor_Apply_MatchingKey(t *testing.T) {
	r, err := NewSecretRedactor([]RedactRule{
		{Pattern: "(?i)password", Replacement: "[REDACTED]"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	input := map[string]string{
		"DB_PASSWORD": "s3cr3t",
		"DB_HOST":     "localhost",
	}
	out := r.Apply(input)
	if out["DB_PASSWORD"] != "[REDACTED]" {
		t.Errorf("expected [REDACTED], got %q", out["DB_PASSWORD"])
	}
	if out["DB_HOST"] != "localhost" {
		t.Errorf("expected localhost, got %q", out["DB_HOST"])
	}
}

func TestSecretRedactor_Apply_DoesNotMutateOriginal(t *testing.T) {
	r, _ := NewSecretRedactor([]RedactRule{
		{Pattern: "SECRET", Replacement: "***"},
	})
	input := map[string]string{"SECRET_KEY": "original"}
	r.Apply(input)
	if input["SECRET_KEY"] != "original" {
		t.Error("original map was mutated")
	}
}

func TestSecretRedactor_Apply_FirstRuleWins(t *testing.T) {
	r, _ := NewSecretRedactor([]RedactRule{
		{Pattern: "TOKEN", Replacement: "[FIRST]"},
		{Pattern: "TOKEN", Replacement: "[SECOND]"},
	})
	out := r.Apply(map[string]string{"API_TOKEN": "abc"})
	if out["API_TOKEN"] != "[FIRST]" {
		t.Errorf("expected [FIRST], got %q", out["API_TOKEN"])
	}
}

func TestRedactStage_InvalidRules_ReturnsError(t *testing.T) {
	stage := RedactStage(nil)
	_, err := stage(map[string]string{"K": "V"})
	if err == nil {
		t.Fatal("expected error from stage with nil rules")
	}
}

func TestRedactStage_ValidRules_RedactsSecrets(t *testing.T) {
	stage := RedactStage([]RedactRule{
		{Pattern: "(?i)secret", Replacement: "<masked>"},
	})
	out, err := stage(map[string]string{"APP_SECRET": "hunter2", "APP_NAME": "vault-sync"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out["APP_SECRET"] != "<masked>" {
		t.Errorf("expected <masked>, got %q", out["APP_SECRET"])
	}
	if out["APP_NAME"] != "vault-sync" {
		t.Errorf("expected vault-sync, got %q", out["APP_NAME"])
	}
}
