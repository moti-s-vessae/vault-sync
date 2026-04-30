package vault_test

import (
	"errors"
	"testing"

	"github.com/your-org/vault-sync/internal/vault"
)

func TestSentinelErrors_AreDistinct(t *testing.T) {
	sentinels := []error{
		vault.ErrAccessDenied,
		vault.ErrSecretNotFound,
		vault.ErrVaultSealed,
		vault.ErrVaultUninitialized,
		vault.ErrCacheExpired,
		vault.ErrInvalidAddress,
	}
	for i := 0; i < len(sentinels); i++ {
		for j := i + 1; j < len(sentinels); j++ {
			if errors.Is(sentinels[i], sentinels[j]) {
				t.Errorf("sentinel errors should be distinct: %v and %v", sentinels[i], sentinels[j])
			}
		}
	}
}

func TestSentinelErrors_MatchWithErrorsIs(t *testing.T) {
	wrapped := func(base error) error {
		return errors.Join(errors.New("context"), base)
	}

	cases := []struct {
		name   string
		err    error
		target error
	}{
		{"access denied", wrapped(vault.ErrAccessDenied), vault.ErrAccessDenied},
		{"not found", wrapped(vault.ErrSecretNotFound), vault.ErrSecretNotFound},
		{"sealed", wrapped(vault.ErrVaultSealed), vault.ErrVaultSealed},
		{"uninitialized", wrapped(vault.ErrVaultUninitialized), vault.ErrVaultUninitialized},
		{"cache expired", wrapped(vault.ErrCacheExpired), vault.ErrCacheExpired},
		{"invalid address", wrapped(vault.ErrInvalidAddress), vault.ErrInvalidAddress},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if !errors.Is(tc.err, tc.target) {
				t.Errorf("expected errors.Is to match %v within wrapped error", tc.target)
			}
		})
	}
}

func TestSentinelErrors_HaveNonEmptyMessages(t *testing.T) {
	sentinels := []error{
		vault.ErrAccessDenied,
		vault.ErrSecretNotFound,
		vault.ErrVaultSealed,
		vault.ErrVaultUninitialized,
		vault.ErrCacheExpired,
		vault.ErrInvalidAddress,
	}
	for _, err := range sentinels {
		if err.Error() == "" {
			t.Errorf("sentinel error should have a non-empty message: %T", err)
		}
	}
}
