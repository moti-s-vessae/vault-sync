package sync

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
)

// SecretHasher computes a deterministic hash over a map of secrets.
// It is used to detect changes between sync cycles without storing plaintext.
type SecretHasher struct{}

// NewSecretHasher returns a new SecretHasher.
func NewSecretHasher() *SecretHasher {
	return &SecretHasher{}
}

// Hash returns a hex-encoded SHA-256 digest of the sorted key=value pairs.
func (h *SecretHasher) Hash(secrets map[string]string) string {
	if len(secrets) == 0 {
		return emptySHA256()
	}

	keys := make([]string, 0, len(secrets))
	for k := range secrets {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	digest := sha256.New()
	for _, k := range keys {
		fmt.Fprintf(digest, "%s=%s\n", k, secrets[k])
	}
	return hex.EncodeToString(digest.Sum(nil))
}

// Equal returns true when both maps produce the same hash.
func (h *SecretHasher) Equal(a, b map[string]string) bool {
	return h.Hash(a) == h.Hash(b)
}

func emptySHA256() string {
	sum := sha256.Sum256([]byte{})
	return hex.EncodeToString(sum[:])
}
