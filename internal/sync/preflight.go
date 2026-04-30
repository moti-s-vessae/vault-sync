package sync

import (
	"context"
	"fmt"
	"time"

	"github.com/user/vault-sync/internal/vault"
)

// PreflightConfig holds configuration for preflight checks.
type PreflightConfig struct {
	VaultAddress string
	Timeout      time.Duration
}

// PreflightChecker runs pre-sync validation checks.
type PreflightChecker struct {
	cfg     PreflightConfig
	checker *vault.HealthChecker
}

// NewPreflightChecker creates a new PreflightChecker.
func NewPreflightChecker(cfg PreflightConfig) *PreflightChecker {
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	return &PreflightChecker{
		cfg:     cfg,
		checker: vault.NewHealthChecker(cfg.VaultAddress, timeout),
	}
}

// Run executes all preflight checks and returns an error if any fail.
func (p *PreflightChecker) Run(ctx context.Context) error {
	if err := p.checkVaultHealth(ctx); err != nil {
		return fmt.Errorf("preflight failed: %w", err)
	}
	return nil
}

func (p *PreflightChecker) checkVaultHealth(ctx context.Context) error {
	status, err := p.checker.Check(ctx)
	if err != nil {
		return fmt.Errorf("vault unreachable at %s: %w", p.cfg.VaultAddress, err)
	}
	if !status.IsReady() {
		if status.Sealed {
			return fmt.Errorf("vault at %s is sealed", p.cfg.VaultAddress)
		}
		if !status.Initialized {
			return fmt.Errorf("vault at %s is not initialized", p.cfg.VaultAddress)
		}
		return fmt.Errorf("vault at %s is not ready", p.cfg.VaultAddress)
	}
	return nil
}
