package sync

import (
	"fmt"

	"github.com/your-org/vault-sync/internal/vault"
)

// MaskedWriter wraps a SecretMasker and provides safe logging of secrets
// by masking values before they are written to any output.
type MaskedWriter struct {
	masker *vault.SecretMasker
}

// NewMaskedWriter creates a MaskedWriter using the provided masking rules.
// Returns an error if any rule pattern is invalid.
func NewMaskedWriter(rules []vault.MaskRule) (*MaskedWriter, error) {
	masker, err := vault.NewSecretMasker(rules)
	if err != nil {
		return nil, fmt.Errorf("masked writer: %w", err)
	}
	return &MaskedWriter{masker: masker}, nil
}

// SafeLog returns a copy of secrets safe for logging, with sensitive values masked.
func (mw *MaskedWriter) SafeLog(secrets map[string]string) map[string]string {
	return mw.masker.MaskSecrets(secrets)
}

// SafeValue returns a single masked value safe for logging.
func (mw *MaskedWriter) SafeValue(key, value string) string {
	return mw.masker.MaskValue(key, value)
}

// DefaultMaskRules returns a sensible default set of masking rules covering
// common sensitive key patterns.
func DefaultMaskRules() []vault.MaskRule {
	return []vault.MaskRule{
		{Pattern: `(?i)password`, Replacement: "[REDACTED]"},
		{Pattern: `(?i)secret`, Replacement: "[REDACTED]"},
		{Pattern: `(?i)token`, Replacement: "[REDACTED]"},
		{Pattern: `(?i)api[_-]?key`, Replacement: "[REDACTED]"},
		{Pattern: `(?i)private[_-]?key`, Replacement: "[REDACTED]"},
	}
}
