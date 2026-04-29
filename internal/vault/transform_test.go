package vault

import (
	"testing"
)

func TestTransformSecrets_BasicTransform(t *testing.T) {
	secrets := map[string]string{
		"app_db_host": "localhost",
		"app_db_port": "5432",
	}
	rules := []TransformRule{
		{Pattern: `^app_`, Replace: ""},
	}
	result, err := TransformSecrets(secrets, rules)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["DB_HOST"] != "localhost" {
		t.Errorf("expected DB_HOST=localhost, got %q", result["DB_HOST"])
	}
	if result["DB_PORT"] != "5432" {
		t.Errorf("expected DB_PORT=5432, got %q", result["DB_PORT"])
	}
}

func TestTransformSecrets_NoRules(t *testing.T) {
	secrets := map[string]string{"foo": "bar"}
	result, err := TransformSecrets(secrets, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["foo"] != "bar" {
		t.Errorf("expected foo=bar, got %v", result)
	}
}

func TestTransformSecrets_MultipleRules(t *testing.T) {
	secrets := map[string]string{
		"service_api_key": "secret",
	}
	rules := []TransformRule{
		{Pattern: `^service_`, Replace: ""},
		{Pattern: `_key$`, Replace: "_token"},
	}
	result, err := TransformSecrets(secrets, rules)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["API_TOKEN"] != "secret" {
		t.Errorf("expected API_TOKEN=secret, got %v", result)
	}
}

func TestTransformSecrets_InvalidPattern(t *testing.T) {
	secrets := map[string]string{"foo": "bar"}
	rules := []TransformRule{
		{Pattern: `[invalid`, Replace: ""},
	}
	_, err := TransformSecrets(secrets, rules)
	if err == nil {
		t.Error("expected error for invalid regex, got nil")
	}
}

func TestTransformSecrets_NoMatchKeepsKey(t *testing.T) {
	secrets := map[string]string{"MY_KEY": "value"}
	rules := []TransformRule{
		{Pattern: `^prefix_`, Replace: ""},
	}
	result, err := TransformSecrets(secrets, rules)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["MY_KEY"] != "value" {
		t.Errorf("expected MY_KEY=value, got %v", result)
	}
}

func TestTransformSecrets_CaptureGroups(t *testing.T) {
	secrets := map[string]string{
		"db_primary_host": "db.example.com",
	}
	rules := []TransformRule{
		{Pattern: `^db_(\w+)_host$`, Replace: "${1}_database_host"},
	}
	result, err := TransformSecrets(secrets, rules)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["PRIMARY_DATABASE_HOST"] != "db.example.com" {
		t.Errorf("expected PRIMARY_DATABASE_HOST, got %v", result)
	}
}
