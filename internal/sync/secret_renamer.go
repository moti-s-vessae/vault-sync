package sync

import (
	"fmt"
	"regexp"
)

// RenameRule maps a regex pattern to a replacement key name template.
type RenameRule struct {
	Pattern     string
	Replacement string
	regexp      *regexp.Regexp
}

// SecretRenamer renames secret keys based on regex rules.
type SecretRenamer struct {
	rules []RenameRule
}

// NewSecretRenamer creates a SecretRenamer from the provided rules.
// Returns an error if any pattern is empty or invalid.
func NewSecretRenamer(rules []RenameRule) (*SecretRenamer, error) {
	if len(rules) == 0 {
		return nil, fmt.Errorf("renamer: at least one rule is required")
	}
	compiled := make([]RenameRule, len(rules))
	for i, r := range rules {
		if r.Pattern == "" {
			return nil, fmt.Errorf("renamer: rule %d has empty pattern", i)
		}
		if r.Replacement == "" {
			return nil, fmt.Errorf("renamer: rule %d has empty replacement", i)
		}
		re, err := regexp.Compile(r.Pattern)
		if err != nil {
			return nil, fmt.Errorf("renamer: rule %d invalid pattern %q: %w", i, r.Pattern, err)
		}
		compiled[i] = RenameRule{Pattern: r.Pattern, Replacement: r.Replacement, regexp: re}
	}
	return &SecretRenamer{rules: compiled}, nil
}

// Apply renames keys in the secrets map according to the first matching rule.
// If a rename would cause a collision, an error is returned.
func (s *SecretRenamer) Apply(secrets map[string]string) (map[string]string, error) {
	out := make(map[string]string, len(secrets))
	for k, v := range secrets {
		newKey := k
		for _, r := range s.rules {
			if r.regexp.MatchString(k) {
				newKey = r.regexp.ReplaceAllString(k, r.Replacement)
				break
			}
		}
		if _, exists := out[newKey]; exists {
			return nil, fmt.Errorf("renamer: key collision on %q", newKey)
		}
		out[newKey] = v
	}
	return out, nil
}

// RenamerStage returns a pipeline Stage that applies the given renamer.
func RenamerStage(r *SecretRenamer) func(map[string]string) (map[string]string, error) {
	return func(secrets map[string]string) (map[string]string, error) {
		return r.Apply(secrets)
	}
}
