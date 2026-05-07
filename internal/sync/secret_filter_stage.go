package sync

import (
	"errors"
	"fmt"
	"regexp"
)

// FilterRule defines a single key-based filter rule applied during pipeline processing.
type FilterRule struct {
	Pattern string `yaml:"pattern"`
	Negate  bool   `yaml:"negate"`
}

// SecretKeyFilter filters secrets by matching keys against a set of compiled regex rules.
type SecretKeyFilter struct {
	rules []compiledFilterRule
}

type compiledFilterRule struct {
	re     *regexp.Regexp
	negate bool
}

// NewSecretKeyFilter compiles the provided FilterRules and returns a SecretKeyFilter.
// Returns an error if any pattern fails to compile.
func NewSecretKeyFilter(rules []FilterRule) (*SecretKeyFilter, error) {
	if len(rules) == 0 {
		return nil, errors.New("secret key filter: at least one rule is required")
	}
	compiled := make([]compiledFilterRule, 0, len(rules))
	for i, r := range rules {
		if r.Pattern == "" {
			return nil, fmt.Errorf("secret key filter: rule[%d] has empty pattern", i)
		}
		re, err := regexp.Compile(r.Pattern)
		if err != nil {
			return nil, fmt.Errorf("secret key filter: rule[%d] invalid pattern %q: %w", i, r.Pattern, err)
		}
		compiled = append(compiled, compiledFilterRule{re: re, negate: r.Negate})
	}
	return &SecretKeyFilter{rules: compiled}, nil
}

// Apply returns a new map containing only the secrets whose keys satisfy all filter rules.
func (f *SecretKeyFilter) Apply(secrets map[string]string) map[string]string {
	out := make(map[string]string, len(secrets))
	for k, v := range secrets {
		if f.matches(k) {
			out[k] = v
		}
	}
	return out
}

func (f *SecretKeyFilter) matches(key string) bool {
	for _, r := range f.rules {
		matched := r.re.MatchString(key)
		if r.negate && matched {
			return false
		}
		if !r.negate && !matched {
			return false
		}
	}
	return true
}

// KeyFilterStage returns a pipeline Stage that applies the given FilterRules to secrets.
func KeyFilterStage(rules []FilterRule) Stage {
	return func(secrets map[string]string) (map[string]string, error) {
		if len(rules) == 0 {
			return secrets, nil
		}
		f, err := NewSecretKeyFilter(rules)
		if err != nil {
			return nil, err
		}
		return f.Apply(secrets), nil
	}
}
