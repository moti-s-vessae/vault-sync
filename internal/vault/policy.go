package vault

import (
	"fmt"
	"strings"
)

// PolicyRule defines a single access rule for a secret path.
type PolicyRule struct {
	Path       string   `yaml:"path"`
	Capabilities []string `yaml:"capabilities"`
}

// Policy represents a set of rules governing secret access.
type Policy struct {
	Rules []PolicyRule `yaml:"rules"`
}

// CheckAccess returns an error if the given path does not have the required capability.
func (p *Policy) CheckAccess(path, capability string) error {
	for _, rule := range p.Rules {
		if matchesPolicy(rule.Path, path) {
			for _, cap := range rule.Capabilities {
				if cap == capability || cap == "*" {
					return nil
				}
			}
			return fmt.Errorf("path %q does not have capability %q", path, capability)
		}
	}
	return fmt.Errorf("no policy rule matches path %q", path)
}

// AllowedPaths returns all paths that have the given capability.
func (p *Policy) AllowedPaths(capability string) []string {
	var paths []string
	for _, rule := range p.Rules {
		for _, cap := range rule.Capabilities {
			if cap == capability || cap == "*" {
				paths = append(paths, rule.Path)
				break
			}
		}
	}
	return paths
}

// matchesPolicy checks if a rule path pattern matches a concrete path.
// Supports trailing wildcard (*) matching.
func matchesPolicy(pattern, path string) bool {
	if pattern == path {
		return true
	}
	if strings.HasSuffix(pattern, "*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(path, prefix)
	}
	return false
}
