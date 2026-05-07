package sync_test

import (
	"testing"

	"github.com/your-org/vault-sync/internal/sync"
)

// staticMultiLoader returns a fixed list of secret maps to simulate multiple
// vault paths being loaded and then deduplicated.
type staticMultiLoader struct {
	maps []map[string]string
}

func (s *staticMultiLoader) Load(_ string) (map[string]string, error) {
	merged := make(map[string]string)
	for _, m := range s.maps {
		for k, v := range m {
			merged[k] = v
		}
	}
	return merged, nil
}

func TestDeduplicateSecrets_Integration_KeepFirst(t *testing.T) {
	sources := []map[string]string{
		{"DB_HOST": "primary.db", "DB_PORT": "5432"},
		{"DB_HOST": "replica.db", "API_KEY": "secret"},
	}

	result, err := sync.DeduplicateSecrets(sources, sync.DeduplicateKeepFirst)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result["DB_HOST"] != "primary.db" {
		t.Errorf("keep-first: expected primary.db, got %q", result["DB_HOST"])
	}
	if result["DB_PORT"] != "5432" {
		t.Errorf("expected DB_PORT to be present")
	}
	if result["API_KEY"] != "secret" {
		t.Errorf("expected API_KEY to be present")
	}
}

func TestDeduplicateSecrets_Integration_KeepLast(t *testing.T) {
	sources := []map[string]string{
		{"TOKEN": "old-token"},
		{"TOKEN": "new-token"},
	}

	result, err := sync.DeduplicateSecrets(sources, sync.DeduplicateKeepLast)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result["TOKEN"] != "new-token" {
		t.Errorf("keep-last: expected new-token, got %q", result["TOKEN"])
	}
}

func TestDeduplicateSecrets_Integration_ErrorStrategy_Propagates(t *testing.T) {
	sources := []map[string]string{
		{"SHARED": "v1"},
		{"SHARED": "v2"},
	}

	_, err := sync.DeduplicateSecrets(sources, sync.DeduplicateError)
	if err == nil {
		t.Error("expected error when duplicate key detected with error strategy")
	}
}
