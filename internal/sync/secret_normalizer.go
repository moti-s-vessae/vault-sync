package sync

import (
	"fmt"
	"regexp"
	"strings"
)

// NormalizeRule defines a key pattern and the normalization strategy to apply.
type NormalizeRule struct {
	Pattern   string
	Strategy  string // "upper", "lower", "snake", "camel"
}

// SecretNormalizer applies key normalization rules to a secret map.
type SecretNormalizer struct {
	rules []compiledNormalizeRule
}

type compiledNormalizeRule struct {
	pattern  *regexp.Regexp
	strategy string
}

// NewSecretNormalizer creates a SecretNormalizer from the provided rules.
// Returns an error if any rule has an empty pattern, unsupported strategy, or invalid regex.
func NewSecretNormalizer(rules []NormalizeRule) (*SecretNormalizer, error) {
	if len(rules) == 0 {
		return nil, fmt.Errorf("normalizer: at least one rule is required")
	}
	compiled := make([]compiledNormalizeRule, 0, len(rules))
	for _, r := range rules {
		if r.Pattern == "" {
			return nil, fmt.Errorf("normalizer: pattern must not be empty")
		}
		switch r.Strategy {
		case "upper", "lower", "snake", "camel":
		default:
			return nil, fmt.Errorf("normalizer: unsupported strategy %q", r.Strategy)
		}
		re, err := regexp.Compile(r.Pattern)
		if err != nil {
			return nil, fmt.Errorf("normalizer: invalid pattern %q: %w", r.Pattern, err)
		}
		compiled = append(compiled, compiledNormalizeRule{pattern: re, strategy: r.Strategy})
	}
	return &SecretNormalizer{rules: compiled}, nil
}

// Apply returns a new map with keys normalized according to the first matching rule.
// Keys with no matching rule are kept as-is.
func (n *SecretNormalizer) Apply(secrets map[string]string) (map[string]string, error) {
	out := make(map[string]string, len(secrets))
	for k, v := range secrets {
		newKey := k
		for _, r := range n.rules {
			if r.pattern.MatchString(k) {
				newKey = applyNormStrategy(k, r.strategy)
				break
			}
		}
		if _, exists := out[newKey]; exists {
			return nil, fmt.Errorf("normalizer: key collision after normalization: %q", newKey)
		}
		out[newKey] = v
	}
	return out, nil
}

func applyNormStrategy(key, strategy string) string {
	switch strategy {
	case "upper":
		return strings.ToUpper(key)
	case "lower":
		return strings.ToLower(key)
	case "snake":
		// Replace hyphens and spaces with underscores, uppercase
		r := strings.NewReplacer("-", "_", " ", "_")
		return strings.ToUpper(r.Replace(key))
	case "camel":
		parts := strings.FieldsFunc(key, func(c rune) bool { return c == '_' || c == '-' || c == ' ' })
		for i, p := range parts {
			if i == 0 {
				parts[i] = strings.ToLower(p)
			} else {
				parts[i] = strings.ToUpper(p[:1]) + strings.ToLower(p[1:])
			}
		}
		return strings.Join(parts, "")
	}
	return key
}

// NormalizeStage returns a pipeline stage that applies the given normalizer.
func NormalizeStage(n *SecretNormalizer) func(map[string]string) (map[string]string, error) {
	return func(secrets map[string]string) (map[string]string, error) {
		return n.Apply(secrets)
	}
}
