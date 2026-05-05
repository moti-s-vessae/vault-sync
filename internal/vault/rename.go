package vault

import "fmt"

// RenameRule maps one secret key to another.
type RenameRule struct {
	From string `yaml:"from"`
	To   string `yaml:"to"`
}

// ApplyRenames renames keys in secrets according to rules.
// First matching rule wins; unmatched keys are kept as-is.
func ApplyRenames(secrets map[string]string, rules []RenameRule) map[string]string {
	if len(rules) == 0 {
		return secrets
	}
	result := make(map[string]string, len(secrets))
	for k, v := range secrets {
		newKey := applyFirstMatch(k, rules)
		result[newKey] = v
	}
	return result
}

// applyFirstMatch returns the renamed key for the first matching rule, or the
// original key if no rule matches.
func applyFirstMatch(key string, rules []RenameRule) string {
	for _, r := range rules {
		if r.From == key {
			return r.To
		}
	}
	return key
}

// ValidateRules checks that no rule has an empty From or To field.
func ValidateRules(rules []RenameRule) error {
	for i, r := range rules {
		if r.From == "" {
			return fmt.Errorf("rename rule[%d]: 'from' must not be empty", i)
		}
		if r.To == "" {
			return fmt.Errorf("rename rule[%d]: 'to' must not be empty", i)
		}
	}
	return nil
}
