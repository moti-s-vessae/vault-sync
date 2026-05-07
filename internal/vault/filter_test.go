package vault

import (
	"testing"
)

func TestFilterSecrets_WithPrefixes(t *testing.T) {
	secrets := map[string]string{
		"APP_DB_HOST":     "localhost",
		"APP_DB_PORT":     "5432",
		"APP_API_KEY":     "secret123",
		"OTHER_SERVICE":   "value",
		"UNRELATED_KEY":   "data",
	}

	prefixes := []string{"APP_DB_", "APP_API_"}
	result := FilterSecrets(secrets, prefixes)

	if len(result) != 3 {
		t.Fatalf("expected 3 secrets, got %d", len(result))
	}
	if result["APP_DB_HOST"] != "localhost" {
		t.Errorf("expected APP_DB_HOST=localhost, got %s", result["APP_DB_HOST"])
	}
	if result["APP_DB_PORT"] != "5432" {
		t.Errorf("expected APP_DB_PORT=5432, got %s", result["APP_DB_PORT"])
	}
	if result["APP_API_KEY"] != "secret123" {
		t.Errorf("expected APP_API_KEY=secret123, got %s", result["APP_API_KEY"])
	}
}

func TestFilterSecrets_EmptyPrefixes(t *testing.T) {
	secrets := map[string]string{
		"APP_KEY": "value1",
		"DB_HOST": "value2",
	}

	result := FilterSecrets(secrets, []string{})

	if len(result) != 2 {
		t.Fatalf("expected all 2 secrets when no prefix filter, got %d", len(result))
	}
}

func TestFilterSecrets_NoMatch(t *testing.T) {
	secrets := map[string]string{
		"APP_KEY": "value1",
		"DB_HOST": "value2",
	}

	result := FilterSecrets(secrets, []string{"NOMATCH_"})

	if len(result) != 0 {
		t.Fatalf("expected 0 secrets, got %d", len(result))
	}
}

func TestFilterSecrets_NilSecrets(t *testing.T) {
	result := FilterSecrets(nil, []string{"APP_"})

	if len(result) != 0 {
		t.Fatalf("expected 0 secrets for nil input, got %d", len(result))
	}
}

func TestStripPrefix_RemovesPrefix(t *testing.T) {
	secrets := map[string]string{
		"APP_DB_HOST": "localhost",
		"APP_DB_PORT": "5432",
		"OTHER_KEY":   "value",
	}

	result := StripPrefix(secrets, "APP_DB_")

	if result["HOST"] != "localhost" {
		t.Errorf("expected HOST=localhost, got %s", result["HOST"])
	}
	if result["PORT"] != "5432" {
		t.Errorf("expected PORT=5432, got %s", result["PORT"])
	}
	if result["OTHER_KEY"] != "value" {
		t.Errorf("expected OTHER_KEY=value to remain unchanged, got %s", result["OTHER_KEY"])
	}
}

func TestMatchesAnyPrefix_True(t *testing.T) {
	if !matchesAnyPrefix("APP_DB_HOST", []string{"OTHER_", "APP_DB_"}) {
		t.Error("expected APP_DB_HOST to match APP_DB_ prefix")
	}
}

func TestMatchesAnyPrefix_False(t *testing.T) {
	if matchesAnyPrefix("APP_DB_HOST", []string{"OTHER_", "NOMATCH_"}) {
		t.Error("expected APP_DB_HOST not to match any prefix")
	}
}
