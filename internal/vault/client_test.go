package vault_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/your-org/vault-sync/internal/vault"
)

func newMockVaultServer(t *testing.T, path string, response map[string]interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != path {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
}

func TestGetSecrets_Success(t *testing.T) {
	payload := map[string]interface{}{
		"data": map[string]interface{}{
			"data": map[string]interface{}{
				"API_KEY": "abc123",
				"DB_PASS": "secret",
			},
		},
	}

	srv := newMockVaultServer(t, "/v1/secret/data/myapp", payload)
	defer srv.Close()

	client, err := vault.NewClient(srv.URL, "test-token", "")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	secrets, err := client.GetSecrets(context.Background(), "secret", "myapp")
	if err != nil {
		t.Fatalf("GetSecrets: %v", err)
	}

	if secrets["API_KEY"] != "abc123" {
		t.Errorf("expected API_KEY=abc123, got %q", secrets["API_KEY"])
	}
	if secrets["DB_PASS"] != "secret" {
		t.Errorf("expected DB_PASS=secret, got %q", secrets["DB_PASS"])
	}
}

func TestGetSecrets_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	client, err := vault.NewClient(srv.URL, "test-token", "")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	_, err = client.GetSecrets(context.Background(), "secret", "missing")
	if err == nil {
		t.Fatal("expected error for missing secret, got nil")
	}
}

func TestNewClient_InvalidAddress(t *testing.T) {
	// Vault SDK accepts any address at construction time; ensure no panic.
	_, err := vault.NewClient("http://localhost:9999", "token", "ns1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
