package vault

import (
	"errors"
	"testing"
	"time"
)

// mockSecretGetter implements SecretGetter for testing.
type mockSecretGetter struct {
	secrets map[string]string
	err     error
	calls   int
}

func (m *mockSecretGetter) GetSecrets(_ string) (map[string]string, error) {
	m.calls++
	return m.secrets, m.err
}

func TestSecretsLoader_Load_Basic(t *testing.T) {
	getter := &mockSecretGetter{
		secrets: map[string]string{
			"app/DB_HOST": "localhost",
			"app/DB_PORT": "5432",
			"other/KEY":  "ignored",
		},
	}
	loader := NewSecretsLoader(getter, nil)
	opts := LoadOptions{Prefixes: []string{"app/"}}

	got, err := loader.Load("secret/myapp", opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got["DB_HOST"] != "localhost" {
		t.Errorf("expected DB_HOST=localhost, got %q", got["DB_HOST"])
	}
	if _, ok := got["other/KEY"]; ok {
		t.Error("expected other/KEY to be filtered out")
	}
}

func TestSecretsLoader_Load_EmptyPath(t *testing.T) {
	loader := NewSecretsLoader(&mockSecretGetter{}, nil)
	_, err := loader.Load("", LoadOptions{})
	if err == nil {
		t.Fatal("expected error for empty path")
	}
}

func TestSecretsLoader_Load_ClientError(t *testing.T) {
	getter := &mockSecretGetter{err: errors.New("vault unavailable")}
	loader := NewSecretsLoader(getter, nil)
	_, err := loader.Load("secret/myapp", LoadOptions{})
	if err == nil {
		t.Fatal("expected error when client fails")
	}
}

func TestSecretsLoader_Load_UsesCache(t *testing.T) {
	getter := &mockSecretGetter{
		secrets: map[string]string{"KEY": "value"},
	}
	cache := NewCache()
	loader := NewSecretsLoader(getter, cache)
	opts := LoadOptions{CacheTTL: 5 * time.Minute}

	// First call populates cache.
	_, err := loader.Load("secret/myapp", opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Second call should hit cache.
	_, err = loader.Load("secret/myapp", opts)
	if err != nil {
		t.Fatalf("unexpected error on second call: %v", err)
	}
	if getter.calls != 1 {
		t.Errorf("expected 1 client call (cache hit), got %d", getter.calls)
	}
}

func TestSecretsLoader_Load_NoCacheTTL_DoesNotCache(t *testing.T) {
	getter := &mockSecretGetter{
		secrets: map[string]string{"KEY": "value"},
	}
	cache := NewCache()
	loader := NewSecretsLoader(getter, cache)
	opts := LoadOptions{} // zero TTL

	loader.Load("secret/myapp", opts) //nolint
	loader.Load("secret/myapp", opts) //nolint
	if getter.calls != 2 {
		t.Errorf("expected 2 client calls (no caching), got %d", getter.calls)
	}
}
