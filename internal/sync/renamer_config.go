package sync

import "fmt"

// RenamerConfig holds configuration for the secret renamer stage.
type RenamerConfig struct {
	Enabled bool         `yaml:"enabled"`
	Rules   []RenameRule `yaml:"rules"`
}

// Validate checks that the config is consistent.
func (c *RenamerConfig) Validate() error {
	if !c.Enabled {
		return nil
	}
	if len(c.Rules) == 0 {
		return fmt.Errorf("renamer: enabled but no rules defined")
	}
	for i, r := range c.Rules {
		if r.Pattern == "" {
			return fmt.Errorf("renamer: rule %d missing pattern", i)
		}
		if r.Replacement == "" {
			return fmt.Errorf("renamer: rule %d missing replacement", i)
		}
	}
	return nil
}

// ToRenamer constructs a SecretRenamer from this config.
// Returns nil, nil when disabled.
func (c *RenamerConfig) ToRenamer() (*SecretRenamer, error) {
	if !c.Enabled {
		return nil, nil
	}
	if err := c.Validate(); err != nil {
		return nil, err
	}
	return NewSecretRenamer(c.Rules)
}

// DefaultRenamerConfig returns a disabled RenamerConfig.
func DefaultRenamerConfig() RenamerConfig {
	return RenamerConfig{Enabled: false}
}
