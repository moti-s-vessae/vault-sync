package sync

import (
	"fmt"
	"unicode/utf8"
)

// TruncateRule defines how a secret value should be truncated.
type TruncateRule struct {
	// Pattern is a glob-style key pattern (e.g. "*_TOKEN", "DB_*").
	Pattern string
	// MaxLen is the maximum number of UTF-8 characters allowed.
	MaxLen int
	// Suffix is appended when truncation occurs (e.g. "..."). Defaults to "".
	Suffix string
}

// SecretTruncator truncates secret values that exceed a configured length.
type SecretTruncator struct {
	rules []TruncateRule
}

// NewSecretTruncator creates a SecretTruncator from the given rules.
// Returns an error if any rule has a non-positive MaxLen or empty Pattern.
func NewSecretTruncator(rules []TruncateRule) (*SecretTruncator, error) {
	if len(rules) == 0 {
		return nil, fmt.Errorf("truncator: at least one rule is required")
	}
	for i, r := range rules {
		if r.Pattern == "" {
			return nil, fmt.Errorf("truncator: rule[%d] has empty pattern", i)
		}
		if r.MaxLen <= 0 {
			return nil, fmt.Errorf("truncator: rule[%d] has non-positive max_len %d", i, r.MaxLen)
		}
	}
	return &SecretTruncator{rules: rules}, nil
}

// Apply returns a new map with values truncated according to the first matching rule.
// Keys with no matching rule are returned unchanged.
func (t *SecretTruncator) Apply(secrets map[string]string) map[string]string {
	out := make(map[string]string, len(secrets))
	for k, v := range secrets {
		out[k] = t.truncate(k, v)
	}
	return out
}

func (t *SecretTruncator) truncate(key, value string) string {
	for _, r := range t.rules {
		if matchGlob(r.Pattern, key) {
			if utf8.RuneCountInString(value) > r.MaxLen {
				runes := []rune(value)
				return string(runes[:r.MaxLen]) + r.Suffix
			}
			return value
		}
	}
	return value
}

// TruncateStage returns a pipeline stage that applies the given truncation rules.
func TruncateStage(rules []TruncateRule) func(map[string]string) (map[string]string, error) {
	return func(secrets map[string]string) (map[string]string, error) {
		tr, err := NewSecretTruncator(rules)
		if err != nil {
			return nil, err
		}
		return tr.Apply(secrets), nil
	}
}

// matchGlob performs simple glob matching supporting only '*' wildcards.
func matchGlob(pattern, s string) bool {
	if pattern == "*" {
		return true
	}
	return globMatch(pattern, s)
}

func globMatch(pat, str string) bool {
	for len(pat) > 0 {
		switch pat[0] {
		case '*':
			if len(pat) == 1 {
				return true
			}
			for i := 0; i <= len(str); i++ {
				if globMatch(pat[1:], str[i:]) {
					return true
				}
			}
			return false
		default:
			if len(str) == 0 || pat[0] != str[0] {
				return false
			}
			pat, str = pat[1:], str[1:]
		}
	}
	return len(str) == 0
}
