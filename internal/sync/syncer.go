package sync

import (
	"context"
	"fmt"
	"log"

	"github.com/example/vault-sync/internal/env"
	"github.com/example/vault-sync/internal/vault"
)

// SecretFetcher abstracts Vault secret retrieval.
type SecretFetcher interface {
	GetSecrets(ctx context.Context, path string) (map[string]string, error)
}

// Syncer orchestrates fetching, filtering, diffing, and writing secrets.
type Syncer struct {
	client  SecretFetcher
	cache   *vault.Cache
	cfg     SyncConfig
	auditor *AuditLogger
}

// SyncConfig holds runtime options for a sync run.
type SyncConfig struct {
	SecretPath string
	OutputFile string
	Prefixes   []string
	Renames    []vault.RenameRule
	UseCache   bool
}

// New creates a Syncer with the provided dependencies.
func New(client SecretFetcher, cache *vault.Cache, cfg SyncConfig, auditor *AuditLogger) *Syncer {
	return &Syncer{client: client, cache: cache, cfg: cfg, auditor: auditor}
}

// Run performs a full sync cycle.
func (s *Syncer) Run(ctx context.Context) error {
	var secrets map[string]string

	if s.cfg.UseCache && s.cache != nil {
		if cached, ok := s.cache.Get(); ok {
			log.Println("vault-sync: using cached secrets")
			secrets = cached
		}
	}

	if secrets == nil {
		var err error
		secrets, err = s.client.GetSecrets(ctx, s.cfg.SecretPath)
		if err != nil {
			return fmt.Errorf("fetching secrets: %w", err)
		}
		if s.cache != nil {
			if err := s.cache.Set(secrets); err != nil {
				log.Printf("vault-sync: warning: could not write cache: %v", err)
			}
		}
	}

	filtered := vault.FilterSecrets(secrets, s.cfg.Prefixes)
	filtered = vault.StripPrefix(filtered, s.cfg.Prefixes)
	filtered = vault.ApplyRenames(filtered, s.cfg.Renames)

	existing, _ := env.ReadEnvFile(s.cfg.OutputFile)
	diff := vault.DiffSecrets(existing, filtered)

	if s.auditor != nil {
		if err := s.auditor.LogChanges(diff); err != nil {
			log.Printf("vault-sync: warning: audit log failed: %v", err)
		}
	}

	if !vault.HasChanges(diff) {
		log.Println("vault-sync: no changes detected")
		return nil
	}

	if err := env.WriteEnvFile(s.cfg.OutputFile, filtered); err != nil {
		return fmt.Errorf("writing env file: %w", err)
	}
	log.Printf("vault-sync: wrote %d secrets to %s", len(filtered), s.cfg.OutputFile)
	return nil
}
