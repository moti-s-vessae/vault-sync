package sync

import "fmt"

// RedactConfig holds configuration for the secret redaction stage,
// typically loaded from .vault-sync.yaml.
type RedactConfig struct {
	Enabled bool          `yaml:"enabled"`
	Rules   []RedactEntry `yaml:"rules"`
}

// RedactEntry maps a yaml-friendly rule definition to a RedactRule.
type RedactEntry struct {
	Pattern     string `yaml:"pattern"`
	Replacement string `yaml:"replacement"`
}

// Validate checks that the RedactConfig is well-formed when enabled.
func (c *RedactConfig) Validate() error {
	if !c.Enabled {
		return nil
	}
	if len(c.Rules) == 0 {
		return fmt.Errorf("redact: enabled but no rules defined")
	}
	for i, r := range c.Rules {
		if r.Pattern == "" {
			return fmt.Errorf("redact: rule[%d] has empty pattern", i)
		}
		if r.Replacement == "" {
			return fmt.Errorf("redact: rule[%d] has empty replacement", i)
		}
	}
	return nil
}

// ToRedactRules converts RedactConfig entries to RedactRule slice.
func (c *RedactConfig) ToRedactRules() []RedactRule {
	rules := make([]RedactRule, 0, len(c.Rules))
	for _, e := range c.Rules {
		rules = append(rules, RedactRule{
			Pattern:     e.Pattern,
			Replacement: e.Replacement,
		})
	}
	return rules
}

// DefaultRedactConfig returns a sensible default RedactConfig
// that redacts common sensitive key patterns.
func DefaultRedactConfig() RedactConfig {
	return RedactConfig{
		Enabled: false,
		Rules: []RedactEntry{
			{Pattern: "(?i)password", Replacement: "[REDACTED]"},
			{Pattern: "(?i)secret", Replacement: "[REDACTED]"},
			{Pattern: "(?i)token", Replacement: "[REDACTED]"},
			{Pattern: "(?i)private_key", Replacement: "[REDACTED]"},
		},
	}
}
