package sync

import (
	"testing"
)

func TestNewSecretKeyFilter_EmptyRules(t *testing.T) {
	_, err := NewSecretKeyFilter(nil)
	if err == nil {
		t.Fatal("expected error for empty rules")
	}
}

func TestNewSecretKeyFilter_InvalidPattern(t *testing.T) {
	_, err := NewSecretKeyFilter([]FilterRule{{Pattern: "[invalid"}})
	if err == nil {
		t.Fatal("expected error for invalid regex pattern")
	}
}

func TestNewSecretKeyFilter_EmptyPattern(t *testing.T) {
	_, err := NewSecretKeyFilter([]FilterRule{{Pattern: ""}})
	if err == nil {
		t.Fatal("expected error for empty pattern")
	}
}

func TestSecretKeyFilter_Apply_MatchingKeys(t *testing.T) {
	f, err := NewSecretKeyFilter([]FilterRule{{Pattern: "^DB_"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	secrets := map[string]string{
		"DB_HOST":    "localhost",
		"DB_PORT":    "5432",
		"APP_SECRET": "abc123",
	}
	result := f.Apply(secrets)
	if len(result) != 2 {
		t.Fatalf("expected 2 results, got %d", len(result))
	}
	if _, ok := result["APP_SECRET"]; ok {
		t.Error("APP_SECRET should have been filtered out")
	}
}

func TestSecretKeyFilter_Apply_NegateRule(t *testing.T) {
	f, err := NewSecretKeyFilter([]FilterRule{{Pattern: "^INTERNAL_", Negate: true}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	secrets := map[string]string{
		"INTERNAL_KEY": "secret",
		"PUBLIC_KEY":   "open",
	}
	result := f.Apply(secrets)
	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}
	if _, ok := result["PUBLIC_KEY"]; !ok {
		t.Error("PUBLIC_KEY should be present")
	}
}

func TestSecretKeyFilter_Apply_EmptySecrets(t *testing.T) {
	f, err := NewSecretKeyFilter([]FilterRule{{Pattern: ".*"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	result := f.Apply(map[string]string{})
	if len(result) != 0 {
		t.Errorf("expected empty result, got %d entries", len(result))
	}
}

func TestKeyFilterStage_NoRules_PassesThrough(t *testing.T) {
	stage := KeyFilterStage(nil)
	secrets := map[string]string{"FOO": "bar"}
	result, err := stage(secrets)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["FOO"] != "bar" {
		t.Error("expected FOO to pass through")
	}
}

func TestKeyFilterStage_InvalidPattern_ReturnsError(t *testing.T) {
	stage := KeyFilterStage([]FilterRule{{Pattern: "[bad"}})
	_, err := stage(map[string]string{"KEY": "val"})
	if err == nil {
		t.Fatal("expected error for invalid pattern")
	}
}
