package vault

// RenameRule describes a single key rename: every secret whose key matches
// From (exact match after optional prefix stripping) is stored under To.
type RenameRule struct {
	From string `yaml:"from"`
	To   string `yaml:"to"`
}

// ApplyRenames returns a new map with rename rules applied. Keys that do not
// match any rule are kept as-is. Rules are evaluated in order; the first
// match wins.
func ApplyRenames(secrets map[string]string, rules []RenameRule) map[string]string {
	if len(rules) == 0 {
		result := make(map[string]string, len(secrets))
		for k, v := range secrets {
			result[k] = v
		}
		return result
	}

	result := make(map[string]string, len(secrets))
	for k, v := range secrets {
		newKey := applyFirstMatch(k, rules)
		result[newKey] = v
	}
	return result
}

func applyFirstMatch(key string, rules []RenameRule) string {
	for _, r := range rules {
		if r.From == key {
			return r.To
		}
	}
	return key
}
