package sync_test

import (
	"testing"

	"github.com/your-org/vault-sync/internal/sync"
)

// TestMaskedWriter_IntegrationWithDefaultRules verifies that a MaskedWriter
// built from DefaultMaskRules correctly masks a realistic secrets map.
func TestMaskedWriter_IntegrationWithDefaultRules(t *testing.T) {
	mw, err := sync.NewMaskedWriter(sync.DefaultMaskRules())
	if err != nil {
		t.Fatalf("failed to create MaskedWriter: %v", err)
	}

	secrets := map[string]string{
		"DB_PASSWORD":      "p@ssw0rd!",
		"APP_SECRET":       "my-app-secret",
		"GITHUB_TOKEN":     "ghp_abc123",
		"STRIPE_API_KEY":   "sk_live_xyz",
		"PRIVATE_KEY":      "-----BEGIN RSA PRIVATE KEY-----",
		"DATABASE_URL":     "postgres://localhost/mydb",
		"LOG_LEVEL":        "info",
		"MAX_CONNECTIONS":  "10",
	}

	safe := mw.SafeLog(secrets)

	redacted := []string{"DB_PASSWORD", "APP_SECRET", "GITHUB_TOKEN", "STRIPE_API_KEY", "PRIVATE_KEY"}
	for _, key := range redacted {
		if safe[key] != "[REDACTED]" {
			t.Errorf("expected %s to be [REDACTED], got %q", key, safe[key])
		}
	}

	plain := []string{"DATABASE_URL", "LOG_LEVEL", "MAX_CONNECTIONS"}
	for _, key := range plain {
		if safe[key] != secrets[key] {
			t.Errorf("expected %s to be unchanged (%q), got %q", key, secrets[key], safe[key])
		}
	}
}
