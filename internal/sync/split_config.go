package sync

import "fmt"

// SplitConfig holds configuration for the secret splitter stage.
type SplitConfig struct {
	// Enabled controls whether the split stage is active.
	Enabled bool `yaml:"enabled"`
	// Rules defines the splitting rules.
	Rules []SplitRule `yaml:"rules"`
}

// Validate checks that the SplitConfig is consistent.
func (c *SplitConfig) Validate() error {
	if !c.Enabled {
		return nil
	}
	if len(c.Rules) == 0 {
		return fmt.Errorf("split: enabled but no rules defined")
	}
	for i, r := range c.Rules {
		if r.Pattern == "" {
			return fmt.Errorf("split: rule[%d] has empty pattern", i)
		}
		if r.Separator == "" {
			return fmt.Errorf("split: rule[%d] has empty separator", i)
		}
	}
	return nil
}

// ToSplitter constructs a SecretSplitter from this config.
// Returns nil, nil if the config is disabled.
func (c *SplitConfig) ToSplitter() (*SecretSplitter, error) {
	if !c.Enabled {
		return nil, nil
	}
	if err := c.Validate(); err != nil {
		return nil, err
	}
	return NewSecretSplitter(c.Rules)
}
