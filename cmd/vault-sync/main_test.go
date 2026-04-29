package main

import (
	"os/exec"
	"strings"
	"testing"
)

// TestMain_Version verifies the --version flag exits cleanly and prints a version string.
func TestMain_Version(t *testing.T) {
	cmd := exec.Command("go", "run", ".", "--version")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("unexpected error: %v\noutput: %s", err, out)
	}
	if !strings.Contains(string(out), "vault-sync") {
		t.Errorf("expected 'vault-sync' in output, got: %s", out)
	}
}

// TestMain_MissingToken verifies the binary exits non-zero when VAULT_TOKEN is absent.
func TestMain_MissingToken(t *testing.T) {
	cmd := exec.Command("go", "run", ".", "--config", "nonexistent.yaml")
	cmd.Env = []string{"HOME=/tmp"} // strip VAULT_TOKEN from env
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected non-zero exit, got success\noutput: %s", out)
	}
}
