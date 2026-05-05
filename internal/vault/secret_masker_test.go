package vault_test

import (
	"testing"

	"github.com/your-org/vault-sync/internal/vault"
)

func TestNewSecretMasker_InvalidPattern(t *testing.T) {
	_, err := vault.NewSecretMasker([]vault.MaskRule{
		{Pattern: "[invalid", Replacement: "***"},
	})
	if err == nil {
		t.Fatal("expected error for invalid regex pattern")
	}
}

func TestMaskSecrets_MatchingKey(t *testing.T) {
	m, err := vault.NewSecretMasker([]vault.MaskRule{
		{Pattern: "(?i)password", Replacement: "[REDACTED]"},
		{Pattern: "(?i)token", Replacement: "[REDACTED]"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	secrets := map[string]string{
		"DB_PASSWORD": "supersecret",
		"API_TOKEN":   "abc123",
		"APP_NAME":    "myapp",
	}

	masked := m.MaskSecrets(secrets)

	if masked["DB_PASSWORD"] != "[REDACTED]" {
		t.Errorf("expected DB_PASSWORD to be redacted, got %q", masked["DB_PASSWORD"])
	}
	if masked["API_TOKEN"] != "[REDACTED]" {
		t.Errorf("expected API_TOKEN to be redacted, got %q", masked["API_TOKEN"])
	}
	if masked["APP_NAME"] != "myapp" {
		t.Errorf("expected APP_NAME to be unchanged, got %q", masked["APP_NAME"])
	}
}

func TestMaskSecrets_DefaultReplacement(t *testing.T) {
	m, err := vault.NewSecretMasker([]vault.MaskRule{
		{Pattern: "SECRET", Replacement: ""},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	masked := m.MaskSecrets(map[string]string{"MY_SECRET": "value"})
	if masked["MY_SECRET"] != "********" {
		t.Errorf("expected default mask, got %q", masked["MY_SECRET"])
	}
}

func TestMaskSecrets_NoRules(t *testing.T) {
	m, err := vault.NewSecretMasker(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	secrets := map[string]string{"KEY": "value"}
	masked := m.MaskSecrets(secrets)
	if masked["KEY"] != "value" {
		t.Errorf("expected value unchanged, got %q", masked["KEY"])
	}
}

func TestMaskSecrets_DoesNotMutateOriginal(t *testing.T) {
	m, _ := vault.NewSecretMasker([]vault.MaskRule{
		{Pattern: "PASS", Replacement: "***"},
	})

	orig := map[string]string{"DB_PASS": "secret"}
	_ = m.MaskSecrets(orig)

	if orig["DB_PASS"] != "secret" {
		t.Error("original map was mutated")
	}
}

func TestMaskValue_SingleKey(t *testing.T) {
	m, _ := vault.NewSecretMasker([]vault.MaskRule{
		{Pattern: "KEY", Replacement: "X"},
	})
	if got := m.MaskValue("MY_KEY", "val"); got != "X" {
		t.Errorf("expected X, got %q", got)
	}
	if got := m.MaskValue("OTHER", "val"); got != "val" {
		t.Errorf("expected val unchanged, got %q", got)
	}
}
