package sync_test

import (
	"testing"

	"github.com/your-org/vault-sync/internal/sync"
	"github.com/your-org/vault-sync/internal/vault"
)

func TestNewMaskedWriter_InvalidPattern(t *testing.T) {
	_, err := sync.NewMaskedWriter([]vault.MaskRule{
		{Pattern: "[bad", Replacement: "***"},
	})
	if err == nil {
		t.Fatal("expected error for invalid pattern")
	}
}

func TestMaskedWriter_SafeLog(t *testing.T) {
	mw, err := sync.NewMaskedWriter(sync.DefaultMaskRules())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	secrets := map[string]string{
		"DB_PASSWORD": "hunter2",
		"API_KEY":     "abc123",
		"APP_ENV":     "production",
	}

	safe := mw.SafeLog(secrets)

	if safe["DB_PASSWORD"] != "[REDACTED]" {
		t.Errorf("DB_PASSWORD should be redacted, got %q", safe["DB_PASSWORD"])
	}
	if safe["API_KEY"] != "[REDACTED]" {
		t.Errorf("API_KEY should be redacted, got %q", safe["API_KEY"])
	}
	if safe["APP_ENV"] != "production" {
		t.Errorf("APP_ENV should be unchanged, got %q", safe["APP_ENV"])
	}
}

func TestMaskedWriter_SafeValue(t *testing.T) {
	mw, err := sync.NewMaskedWriter(sync.DefaultMaskRules())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := mw.SafeValue("AUTH_TOKEN", "tok_secret"); got != "[REDACTED]" {
		t.Errorf("expected [REDACTED], got %q", got)
	}
	if got := mw.SafeValue("REGION", "us-east-1"); got != "us-east-1" {
		t.Errorf("expected us-east-1, got %q", got)
	}
}

func TestMaskedWriter_DoesNotMutateInput(t *testing.T) {
	mw, _ := sync.NewMaskedWriter(sync.DefaultMaskRules())

	orig := map[string]string{"DB_PASSWORD": "secret"}
	_ = mw.SafeLog(orig)

	if orig["DB_PASSWORD"] != "secret" {
		t.Error("original map was mutated by SafeLog")
	}
}

func TestDefaultMaskRules_NotEmpty(t *testing.T) {
	rules := sync.DefaultMaskRules()
	if len(rules) == 0 {
		t.Error("expected at least one default mask rule")
	}
	for _, r := range rules {
		if r.Pattern == "" {
			t.Error("default rule has empty pattern")
		}
	}
}
