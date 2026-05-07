package sync

import (
	"fmt"
	"log"

	"github.com/yourusername/vault-sync/internal/vault"
)

// PolicyGuard enforces a vault Policy before syncing secrets from a path.
type PolicyGuard struct {
	policy     *vault.Policy
	capability string
	logger     *log.Logger
}

// NewPolicyGuard creates a PolicyGuard that checks the given capability.
func NewPolicyGuard(policy *vault.Policy, capability string, logger *log.Logger) *PolicyGuard {
	if logger == nil {
		logger = log.Default()
	}
	return &PolicyGuard{
		policy:     policy,
		capability: capability,
		logger:     logger,
	}
}

// Check returns an error if any of the provided paths are not permitted.
func (g *PolicyGuard) Check(paths []string) error {
	for _, path := range paths {
		if err := g.policy.CheckAccess(path, g.capability); err != nil {
			g.logger.Printf("[policy] access denied: %v", err)
			return fmt.Errorf("policy check failed for %q: %w", path, err)
		}
		g.logger.Printf("[policy] access granted: path=%q capability=%q", path, g.capability)
	}
	return nil
}

// FilterAllowed returns only the paths permitted by the policy.
func (g *PolicyGuard) FilterAllowed(paths []string) []string {
	var allowed []string
	for _, path := range paths {
		if err := g.policy.CheckAccess(path, g.capability); err == nil {
			allowed = append(allowed, path)
		}
	}
	return allowed
}

// DeniedPaths returns the subset of paths that are NOT permitted by the policy.
// This is useful for audit logging or reporting which paths would be blocked.
func (g *PolicyGuard) DeniedPaths(paths []string) []string {
	var denied []string
	for _, path := range paths {
		if err := g.policy.CheckAccess(path, g.capability); err != nil {
			g.logger.Printf("[policy] denied path detected: path=%q capability=%q", path, g.capability)
			denied = append(denied, path)
		}
	}
	return denied
}
