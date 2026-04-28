package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_DefaultsAndEnvVars(t *testing.T) {
	t.Setenv("VAULT_TOKEN", "test-token")
	t.Setenv("VAULT_ADDR", "http://vault.example.com:8200")

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.VaultToken != "test-token" {
		t.Errorf("expected token %q, got %q", "test-token", cfg.VaultToken)
	}
	if cfg.VaultAddr != "http://vault.example.com:8200" {
		t.Errorf("expected addr %q, got %q", "http://vault.example.com:8200", cfg.VaultAddr)
	}
	if cfg.OutputFile != ".env" {
		t.Errorf("expected default output file %q, got %q", ".env", cfg.OutputFile)
	}
	if cfg.MountPath != "secret" {
		t.Errorf("expected default mount path %q, got %q", "secret", cfg.MountPath)
	}
}

func TestLoad_MissingToken(t *testing.T) {
	os.Unsetenv("VAULT_TOKEN")
	os.Unsetenv("VAULT_SYNC_VAULT_TOKEN")

	_, err := Load("")
	if err == nil {
		t.Fatal("expected error when vault token is missing, got nil")
	}
}

func TestLoad_FromFile(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")

	content := []byte(`
vault_addr: "http://localhost:9200"
vault_token: "file-token"
namespace: "myapp/prod"
output_file: "secrets.env"
mount_path: "kv"
`)
	if err := os.WriteFile(cfgPath, content, 0o600); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}

	// Ensure env vars don't interfere.
	os.Unsetenv("VAULT_TOKEN")
	os.Unsetenv("VAULT_ADDR")

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.VaultAddr != "http://localhost:9200" {
		t.Errorf("expected addr from file, got %q", cfg.VaultAddr)
	}
	if cfg.Namespace != "myapp/prod" {
		t.Errorf("expected namespace %q, got %q", "myapp/prod", cfg.Namespace)
	}
	if cfg.OutputFile != "secrets.env" {
		t.Errorf("expected output file %q, got %q", "secrets.env", cfg.OutputFile)
	}
	if cfg.MountPath != "kv" {
		t.Errorf("expected mount path %q, got %q", "kv", cfg.MountPath)
	}
}
