package sync

import (
	"strings"
	"testing"
)

func TestNewSecretLimiter_NoRules(t *testing.T) {
	_, err := NewSecretLimiter(nil)
	if err == nil {
		t.Fatal("expected error for empty rules")
	}
}

func TestNewSecretLimiter_EmptyPattern(t *testing.T) {
	_, err := NewSecretLimiter([]LimitRule{{Pattern: "", MaxCount: 5}})
	if err == nil || !strings.Contains(err.Error(), "empty pattern") {
		t.Fatalf("expected empty pattern error, got %v", err)
	}
}

func TestNewSecretLimiter_NonPositiveMaxCount(t *testing.T) {
	_, err := NewSecretLimiter([]LimitRule{{Pattern: "APP_*", MaxCount: 0}})
	if err == nil || !strings.Contains(err.Error(), "max_count must be positive") {
		t.Fatalf("expected max_count error, got %v", err)
	}
}

func TestSecretLimiter_Apply_WithinLimit(t *testing.T) {
	limiter, err := NewSecretLimiter([]LimitRule{{Pattern: "APP_*", MaxCount: 3}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	secrets := map[string]string{
		"APP_FOO": "1",
		"APP_BAR": "2",
		"DB_HOST": "localhost",
	}
	out, err := limiter.Apply(secrets)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out) != 3 {
		t.Errorf("expected 3 secrets, got %d", len(out))
	}
}

func TestSecretLimiter_Apply_ExceedsLimit(t *testing.T) {
	limiter, err := NewSecretLimiter([]LimitRule{{Pattern: "APP_*", MaxCount: 1}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	secrets := map[string]string{
		"APP_FOO": "1",
		"APP_BAR": "2",
	}
	_, err = limiter.Apply(secrets)
	if err == nil || !strings.Contains(err.Error(), "exceeds max_count") {
		t.Fatalf("expected limit exceeded error, got %v", err)
	}
}

func TestSecretLimiter_Apply_DoesNotMutateInput(t *testing.T) {
	limiter, _ := NewSecretLimiter([]LimitRule{{Pattern: "*", MaxCount: 10}})
	input := map[string]string{"KEY": "value"}
	out, _ := limiter.Apply(input)
	out["KEY"] = "mutated"
	if input["KEY"] != "value" {
		t.Error("Apply mutated the input map")
	}
}

func TestLimitStage_Integration_WithPipeline(t *testing.T) {
	rules := []LimitRule{{Pattern: "SECRET_*", MaxCount: 2}}
	secrets := map[string]string{
		"SECRET_A": "a",
		"SECRET_B": "b",
		"SECRET_C": "c",
		"OTHER":    "x",
	}
	stage := LimitStage(rules)
	_, err := stage(secrets)
	if err == nil || !strings.Contains(err.Error(), "exceeds max_count") {
		t.Fatalf("expected limit error from stage, got %v", err)
	}
}

func TestLimitStage_Integration_NoViolation(t *testing.T) {
	rules := []LimitRule{{Pattern: "SECRET_*", MaxCount: 5}}
	secrets := map[string]string{
		"SECRET_A": "a",
		"SECRET_B": "b",
	}
	stage := LimitStage(rules)
	out, err := stage(secrets)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out) != 2 {
		t.Errorf("expected 2 secrets, got %d", len(out))
	}
}
