package sync_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/example/vault-sync/internal/config"
	vsync "github.com/example/vault-sync/internal/sync"
)

func newMockVault(t *testing.T, data map[string]interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{"data": data},
		})
	}))
}

func TestRun_WritesFilteredSecrets(t *testing.T) {
	server := newMockVault(t, map[string]interface{}{
		"APP_KEY": "abc123",
		"DB_PASS": "secret",
		"OTHER":   "ignore",
	})
	defer server.Close()

	out := filepath.Join(t.TempDir(), ".env")
	cfg := &config.Config{
		VaultAddr:   server.URL,
		VaultToken:  "test-token",
		SecretPath:  "secret/data/myapp",
		Prefixes:    []string{"APP_", "DB_"},
		StripPrefix: false,
		OutputFile:  out,
	}

	s, err := vsync.New(cfg)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	res, err := s.Run()
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	if res.SecretsTotal != 3 {
		t.Errorf("SecretsTotal = %d, want 3", res.SecretsTotal)
	}
	if res.SecretsWritten != 2 {
		t.Errorf("SecretsWritten = %d, want 2", res.SecretsWritten)
	}
	if res.OutputFile != out {
		t.Errorf("OutputFile = %q, want %q", res.OutputFile, out)
	}

	if _, err := os.Stat(out); err != nil {
		t.Errorf("output file not created: %v", err)
	}
}

func TestRun_NoFilter_WritesAll(t *testing.T) {
	server := newMockVault(t, map[string]interface{}{
		"FOO": "bar",
		"BAZ": "qux",
	})
	defer server.Close()

	out := filepath.Join(t.TempDir(), ".env")
	cfg := &config.Config{
		VaultAddr:  server.URL,
		VaultToken: "test-token",
		SecretPath: "secret/data/myapp",
		OutputFile: out,
	}

	s, err := vsync.New(cfg)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	res, err := s.Run()
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	if res.SecretsWritten != 2 {
		t.Errorf("SecretsWritten = %d, want 2", res.SecretsWritten)
	}
}
