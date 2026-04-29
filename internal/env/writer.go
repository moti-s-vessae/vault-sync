package env

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

// WriteEnvFile writes a map of key-value secret pairs to a .env file at the given path.
// Existing files will be overwritten. Keys are written in sorted order for determinism.
func WriteEnvFile(path string, secrets map[string]string) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to open env file %q: %w", path, err)
	}
	defer file.Close()

	keys := make([]string, 0, len(secrets))
	for k := range secrets {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		line := fmt.Sprintf("%s=%s\n", k, quoteValue(secrets[k]))
		if _, err := file.WriteString(line); err != nil {
			return fmt.Errorf("failed to write key %q: %w", k, err)
		}
	}

	return nil
}

// quoteValue wraps values containing spaces or special characters in double quotes.
func quoteValue(v string) string {
	if strings.ContainsAny(v, " \t\n#\"\'\\$") {
		escaped := strings.ReplaceAll(v, `"`, `\"`)
		return `"` + escaped + `"`
	}
	return v
}
