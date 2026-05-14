package sync

import "fmt"

// AnnotateConfig holds configuration for the annotation stage loaded from YAML.
type AnnotateConfig struct {
	Enabled bool             `yaml:"enabled"`
	Rules   []AnnotationRule `yaml:"rules"`
}

// Validate checks that the configuration is consistent.
// It returns an error if the config is enabled but has no rules, or if any
// rule is missing a required field (pattern or tag_key).
func (c *AnnotateConfig) Validate() error {
	if !c.Enabled {
		return nil
	}
	if len(c.Rules) == 0 {
		return fmt.Errorf("annotate: enabled but no rules defined")
	}
	for i, r := range c.Rules {
		if r.Pattern == "" {
			return fmt.Errorf("annotate: rule[%d] has empty pattern", i)
		}
		if r.TagKey == "" {
			return fmt.Errorf("annotate: rule[%d] has empty tag_key", i)
		}
	}
	return nil
}

// ToAnnotator builds a SecretAnnotator from the config.
// Returns nil, nil when the feature is disabled.
func (c *AnnotateConfig) ToAnnotator() (*SecretAnnotator, error) {
	if !c.Enabled {
		return nil, nil
	}
	if err := c.Validate(); err != nil {
		return nil, err
	}
	return NewSecretAnnotator(c.Rules)
}
