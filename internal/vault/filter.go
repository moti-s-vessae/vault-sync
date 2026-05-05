package vault

import "strings"

// FilterSecrets returns only the secrets whose keys match at least one of the
// given prefixes. If prefixes is empty every secret is returned unchanged.
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

// StripPrefix removes the first matching prefix from key. If no prefix
// matches the original key is returned unchanged.
func StripPrefix(key string, prefixes []string) string {
	for _, p := range prefixes {
		if strings.HasPrefix(key, p) {
			return strings.TrimPrefix(key, p)
		}
	}
	return key
}

// StripPrefixFromAll returns a new map with StripPrefix applied to every key.
// Values are preserved unchanged. If two keys reduce to the same stripped key
// the last one processed wins (map iteration order is not guaranteed).
func StripPrefixFromAll(secrets map[string]string, prefixes []string) map[string]string {
	result := make(map[string]string, len(secrets))
	for k, v := range secrets {
		result[StripPrefix(k, prefixes)] = v
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
