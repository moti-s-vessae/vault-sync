package sync

import "fmt"

// AnnotateConfig holds configuration for the annotation stage loaded from YAML.
type AnnotateConfig struct {
	Enabled bool             `yaml:"enabled"`
	Rules   []AnnotationRule `yaml:"rules"`
}

// Validate checks that the configuration is consistent.
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
	return NewSecretAnnotator(c.Rules)
}
