package sync

import (
	"fmt"
	"sort"
)

// LimitRule defines a maximum number of secrets allowed for a given glob pattern.
type LimitRule struct {
	Pattern  string
	MaxCount int
}

// SecretLimiter enforces per-group maximum secret count rules.
type SecretLimiter struct {
	rules []LimitRule
}

// NewSecretLimiter creates a SecretLimiter from the provided rules.
// Returns an error if any rule has an empty pattern or a non-positive MaxCount.
func NewSecretLimiter(rules []LimitRule) (*SecretLimiter, error) {
	if len(rules) == 0 {
		return nil, fmt.Errorf("secret limiter: at least one rule is required")
	}
	for i, r := range rules {
		if r.Pattern == "" {
			return nil, fmt.Errorf("secret limiter: rule[%d] has empty pattern", i)
		}
		if r.MaxCount <= 0 {
			return nil, fmt.Errorf("secret limiter: rule[%d] max_count must be positive, got %d", i, r.MaxCount)
		}
	}
	return &SecretLimiter{rules: rules}, nil
}

// Apply enforces limits on secrets, returning an error if any rule is exceeded.
// Secrets are evaluated in sorted key order so truncation is deterministic.
func (l *SecretLimiter) Apply(secrets map[string]string) (map[string]string, error) {
	keys := make([]string, 0, len(secrets))
	for k := range secrets {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, rule := range l.rules {
		var matched []string
		for _, k := range keys {
			if matchGlob(rule.Pattern, k) {
				matched = append(matched, k)
			}
		}
		if len(matched) > rule.MaxCount {
			return nil, fmt.Errorf(
				"secret limiter: pattern %q matched %d secrets, exceeds max_count of %d",
				rule.Pattern, len(matched), rule.MaxCount,
			)
		}
	}

	out := make(map[string]string, len(secrets))
	for k, v := range secrets {
		out[k] = v
	}
	return out, nil
}

// LimitStage returns a pipeline stage that enforces secret count limits.
func LimitStage(rules []LimitRule) func(map[string]string) (map[string]string, error) {
	return func(secrets map[string]string) (map[string]string, error) {
		limiter, err := NewSecretLimiter(rules)
		if err != nil {
			return nil, err
		}
		return limiter.Apply(secrets)
	}
}
