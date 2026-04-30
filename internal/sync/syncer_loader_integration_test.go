package sync_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/your-org/vault-sync/internal/vault"
)

// fakeClient implements vault.SecretGetter for integration tests.
type fakeClient struct {
	data map[string]string
}

func (f *fakeClient) GetSecrets(_ string) (map[string]string, error) {
	return f.data, nil
}

func TestSecretsLoader_IntegrationWithCache(t *testing.T) {
	client := &fakeClient{
		data: map[string]string{
			"prod/API_KEY":    "secret-key",
			"prod/DB_PASS":    "db-pass",
			"staging/API_KEY": "staging-key",
		},
	}
	cache := vault.NewCache()
	loader := vault.NewSecretsLoader(client, cache)

	opts := vault.LoadOptions{
		Prefixes: []string{"prod/"},
		CacheTTL: 10 * time.Minute,
	}

	got, err := loader.Load("secret/env", opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got["API_KEY"] != "secret-key" {
		t.Errorf("expected API_KEY=secret-key, got %q", got["API_KEY"])
	}
	if _, ok := got["staging/API_KEY"]; ok {
		t.Error("staging key should have been filtered")
	}
}

func TestSecretsLoader_WritesExpectedEnvFile(t *testing.T) {
	client := &fakeClient{
		data: map[string]string{
			"app/HOST": "localhost",
			"app/PORT": "8080",
		},
	}
	loader := vault.NewSecretsLoader(client, nil)
	opts := vault.LoadOptions{Prefixes: []string{"app/"}}

	secrets, err := loader.Load("secret/svc", opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, ".env")

	f, err := os.Create(outPath)
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	for k, v := range secrets {
		f.WriteString(k + "=" + v + "\n")
	}
	f.Close()

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}
	contents := string(data)
	if len(contents) == 0 {
		t.Error("expected non-empty .env file")
	}
}
