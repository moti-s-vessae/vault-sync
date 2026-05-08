package sync

import (
	"fmt"
	"regexp"
)

// RedactRule defines a pattern and the replacement string for matching keys.
type RedactRule struct {
	Pattern     string
	Replacement string
}

// SecretRedactor replaces values of matching keys with a redacted placeholder.
type SecretRedactor struct {
	rules []compiledRedactRule
}

type compiledRedactRule struct {
	re          *regexp.Regexp
	replacement string
}

// NewSecretRedactor compiles the given rules and returns a SecretRedactor.
// Returns an error if any pattern is invalid or replacement is empty.
func NewSecretRedactor(rules []RedactRule) (*SecretRedactor, error) {
	if len(rules) == 0 {
		return nil, fmt.Errorf("redactor: at least one rule is required")
	}
	compiled := make([]compiledRedactRule, 0, len(rules))
	for _, r := range rules {
		if r.Pattern == "" {
			return nil, fmt.Errorf("redactor: pattern must not be empty")
		}
		if r.Replacement == "" {
			return nil, fmt.Errorf("redactor: replacement must not be empty for pattern %q", r.Pattern)
		}
		re, err := regexp.Compile(r.Pattern)
		if err != nil {
			return nil, fmt.Errorf("redactor: invalid pattern %q: %w", r.Pattern, err)
		}
		compiled = append(compiled, compiledRedactRule{re: re, replacement: r.Replacement})
	}
	return &SecretRedactor{rules: compiled}, nil
}

// Apply returns a new map with values redacted for keys matching any rule.
// The original map is never mutated.
func (r *SecretRedactor) Apply(secrets map[string]string) map[string]string {
	out := make(map[string]string, len(secrets))
	for k, v := range secrets {
		out[k] = v
	}
	for k := range out {
		for _, rule := range r.rules {
			if rule.re.MatchString(k) {
				out[k] = rule.replacement
				break
			}
		}
	}
	return out
}

// RedactStage returns a pipeline stage that redacts secrets using the given rules.
// Returns an error stage if rule compilation fails.
func RedactStage(rules []RedactRule) func(map[string]string) (map[string]string, error) {
	return func(secrets map[string]string) (map[string]string, error) {
		redactor, err := NewSecretRedactor(rules)
		if err != nil {
			return nil, err
		}
		return redactor.Apply(secrets), nil
	}
}
