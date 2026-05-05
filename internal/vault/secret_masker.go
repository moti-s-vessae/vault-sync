package vault

import (
	"regexp"
	"strings"
)

// MaskRule defines a pattern and its replacement for masking secret values.
type MaskRule struct {
	Pattern     string
	Replacement string
}

// SecretMasker masks sensitive values in a secrets map before logging or display.
type SecretMasker struct {
	rules    []MaskRule
	compiled []*regexp.Regexp
}

// NewSecretMasker creates a SecretMasker from the provided rules.
// Returns an error if any pattern fails to compile.
func NewSecretMasker(rules []MaskRule) (*SecretMasker, error) {
	compiled := make([]*regexp.Regexp, 0, len(rules))
	for _, r := range rules {
		re, err := regexp.Compile(r.Pattern)
		if err != nil {
			return nil, err
		}
		compiled = append(compiled, re)
	}
	return &SecretMasker{rules: rules, compiled: compiled}, nil
}

// MaskSecrets returns a copy of secrets with values masked according to rules.
// Keys matching a pattern have their values replaced with the corresponding replacement.
func (m *SecretMasker) MaskSecrets(secrets map[string]string) map[string]string {
	result := make(map[string]string, len(secrets))
	for k, v := range secrets {
		result[k] = m.maskValue(k, v)
	}
	return result
}

// MaskValue masks a single value if its key matches any rule pattern.
func (m *SecretMasker) MaskValue(key, value string) string {
	return m.maskValue(key, value)
}

func (m *SecretMasker) maskValue(key, value string) string {
	for i, re := range m.compiled {
		if re.MatchString(key) {
			replacement := m.rules[i].Replacement
			if replacement == "" {
				replacement = strings.Repeat("*", 8)
			}
			return replacement
		}
	}
	return value
}
