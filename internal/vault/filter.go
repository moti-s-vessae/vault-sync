package vault

import (
	"strings"
)

// FilterSecrets returns a new map containing only the entries whose keys
// match at least one of the provided namespace prefixes.
// If prefixes is empty, all entries are returned unchanged.
func FilterSecrets(secrets map[string]string, prefixes []string) map[string]string {
	if len(prefixes) == 0 {
		return secrets
	}

	filtered := make(map[string]string)
	for k, v := range secrets {
		if matchesAnyPrefix(k, prefixes) {
			filtered[k] = v
		}
	}

	return filtered
}

// StripPrefix removes the given prefix from all matching keys in the map,
// returning a new map with the modified keys. Non-matching keys are kept as-is.
func StripPrefix(secrets map[string]string, prefix string) map[string]string {
	if prefix == "" {
		return secrets
	}

	result := make(map[string]string, len(secrets))
	for k, v := range secrets {
		if strings.HasPrefix(k, prefix) {
			result[strings.TrimPrefix(k, prefix)] = v
		} else {
			result[k] = v
		}
	}

	return result
}

func matchesAnyPrefix(key string, prefixes []string) bool {
	for _, p := range prefixes {
		if strings.HasPrefix(key, p) {
			return true
		}
	}
	return false
}
