package vault

import (
	"testing"
)

func TestApplyRenames_BasicRename(t *testing.T) {
	secrets := map[string]string{
		"APP_DB_HOST": "localhost",
		"APP_DB_PORT": "5432",
		"APP_SECRET":  "s3cr3t",
	}
	rules := []RenameRule{
		{From: "APP_DB_HOST", To: "DATABASE_HOST"},
		{From: "APP_DB_PORT", To: "DATABASE_PORT"},
	}

	got := ApplyRenames(secrets, rules)

	if got["DATABASE_HOST"] != "localhost" {
		t.Errorf("expected DATABASE_HOST=localhost, got %q", got["DATABASE_HOST"])
	}
	if got["DATABASE_PORT"] != "5432" {
		t.Errorf("expected DATABASE_PORT=5432, got %q", got["DATABASE_PORT"])
	}
	if got["APP_SECRET"] != "s3cr3t" {
		t.Errorf("expected APP_SECRET=s3cr3t, got %q", got["APP_SECRET"])
	}
	if _, exists := got["APP_DB_HOST"]; exists {
		t.Error("old key APP_DB_HOST should not exist after rename")
	}
}

func TestApplyRenames_NoRules(t *testing.T) {
	secrets := map[string]string{"FOO": "bar", "BAZ": "qux"}
	got := ApplyRenames(secrets, nil)

	if len(got) != len(secrets) {
		t.Fatalf("expected %d keys, got %d", len(secrets), len(got))
	}
	for k, v := range secrets {
		if got[k] != v {
			t.Errorf("key %s: expected %q, got %q", k, v, got[k])
		}
	}
}

func TestApplyRenames_FirstMatchWins(t *testing.T) {
	secrets := map[string]string{"KEY": "value"}
	rules := []RenameRule{
		{From: "KEY", To: "FIRST"},
		{From: "KEY", To: "SECOND"},
	}

	got := ApplyRenames(secrets, rules)

	if got["FIRST"] != "value" {
		t.Errorf("expected FIRST=value (first rule wins), got %v", got)
	}
	if _, exists := got["SECOND"]; exists {
		t.Error("SECOND should not exist; first rule should have matched")
	}
}

func TestApplyRenames_NoMatchKeepsOriginal(t *testing.T) {
	secrets := map[string]string{"UNMATCHED": "val"}
	rules := []RenameRule{{From: "OTHER", To: "NEW"}}

	got := ApplyRenames(secrets, rules)

	if got["UNMATCHED"] != "val" {
		t.Errorf("expected UNMATCHED=val, got %q", got["UNMATCHED"])
	}
}
