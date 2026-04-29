package vault

import "strings"

// FilterSecrets returns only the secrets whose keys match at least one of the
// given prefixes. If prefixes is empty, all secrets are returned unchanged.
func FilterSecrets(secrets map[string]string, prefixes []string) map[string]string {
	if len(prefixes) == 0 {
		result := make(map[string]string, len(secrets))
		for k, v := range secrets {
			result[k] = v
		}
		return result
	}

	result := make(map[string]string)
	for k, v := range secrets {
		if matchesAnyPrefix(k, prefixes) {
			result[k] = v
		}
	}
	return result
}

// StripPrefix removes the given prefix from all keys that start with it.
// Keys that do not start with the prefix are kept as-is.
func StripPrefix(secrets map[string]string, prefix string) map[string]string {
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

// matchesAnyPrefix reports whether key starts with at least one of the given prefixes.
func matchesAnyPrefix(key string, prefixes []string) bool {
	for _, p := range prefixes {
		if strings.HasPrefix(key, p) {
			return true
		}
	}
	return false
}
