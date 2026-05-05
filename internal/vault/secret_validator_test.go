package vault

import (
	"testing"
)

func TestValidateSecrets_NoRules(t *testing.T) {
	secrets := map[string]string{"FOO": "bar"}
	if err := ValidateSecrets(secrets, nil); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestValidateSecrets_PatternMatch(t *testing.T) {
	secrets := map[string]string{
		"DB_PASSWORD": "s3cr3t!",
		"DB_HOST":     "localhost",
	}
	rules := []ValidationRule{
		{Key: "^DB_HOST$", Pattern: `^[a-z]+$`},
	}
	if err := ValidateSecrets(secrets, rules); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestValidateSecrets_PatternMismatch(t *testing.T) {
	secrets := map[string]string{
		"API_KEY": "not-hex!",
	}
	rules := []ValidationRule{
		{Key: "^API_KEY$", Pattern: `^[0-9a-f]{32}$`},
	}
	err := ValidateSecrets(secrets, rules)
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
	if !IsValidationError(err) {
		t.Fatalf("expected *ValidationError, got %T", err)
	}
}

func TestValidateSecrets_RequiredKeyPresent(t *testing.T) {
	secrets := map[string]string{"REQUIRED_KEY": "value"}
	rules := []ValidationRule{
		{Key: "^REQUIRED_KEY$", Required: true},
	}
	if err := ValidateSecrets(secrets, rules); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestValidateSecrets_RequiredKeyMissing(t *testing.T) {
	secrets := map[string]string{"OTHER_KEY": "value"}
	rules := []ValidationRule{
		{Key: "^REQUIRED_KEY$", Required: true},
	}
	err := ValidateSecrets(secrets, rules)
	if err == nil {
		t.Fatal("expected error for missing required key")
	}
	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("expected *ValidationError")
	}
	if len(ve.Violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(ve.Violations))
	}
}

func TestValidateSecrets_InvalidKeyPattern(t *testing.T) {
	secrets := map[string]string{"FOO": "bar"}
	rules := []ValidationRule{
		{Key: "[invalid"},
	}
	err := ValidateSecrets(secrets, rules)
	if err == nil {
		t.Fatal("expected error for invalid key pattern")
	}
}

func TestValidateSecrets_MultipleViolations(t *testing.T) {
	secrets := map[string]string{
		"PORT": "not-a-number",
	}
	rules := []ValidationRule{
		{Key: "^PORT$", Pattern: `^\d+$`},
		{Key: "^MISSING$", Required: true},
	}
	err := ValidateSecrets(secrets, rules)
	if err == nil {
		t.Fatal("expected error")
	}
	ve := err.(*ValidationError)
	if len(ve.Violations) != 2 {
		t.Fatalf("expected 2 violations, got %d: %v", len(ve.Violations), ve.Violations)
	}
}

func TestIsValidationError_False(t *testing.T) {
	if IsValidationError(nil) {
		t.Fatal("nil should not be a ValidationError")
	}
}
