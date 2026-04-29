package sync

import (
	"context"
	"fmt"
	"log"

	"github.com/user/vault-sync/internal/env"
	"github.com/user/vault-sync/internal/vault"
)

// VaultClient abstracts Vault secret retrieval.
type VaultClient interface {
	GetSecrets(ctx context.Context, path string) (map[string]string, error)
}

// Syncer orchestrates fetching, filtering, diffing, and writing secrets.
type Syncer struct {
	client   VaultClient
	path     string
	output   string
	prefixes []string
	renames  []vault.RenameRule
	dryRun   bool
}

// New creates a new Syncer.
func New(client VaultClient, path, output string, prefixes []string, renames []vault.RenameRule, dryRun bool) *Syncer {
	return &Syncer{
		client:   client,
		path:     path,
		output:   output,
		prefixes: prefixes,
		renames:  renames,
		dryRun:   dryRun,
	}
}

// Run executes the sync: fetch → filter → rename → diff → write.
func (s *Syncer) Run(ctx context.Context) error {
	secrets, err := s.client.GetSecrets(ctx, s.path)
	if err != nil {
		return fmt.Errorf("fetching secrets: %w", err)
	}

	if len(s.prefixes) > 0 {
		secrets = vault.FilterSecrets(secrets, s.prefixes)
		secrets = vault.StripPrefix(secrets, s.prefixes)
	}

	if len(s.renames) > 0 {
		secrets = vault.ApplyRenames(secrets, s.renames)
	}

	existing, err := env.ReadEnvFile(s.output)
	if err != nil {
		return fmt.Errorf("reading existing env file: %w", err)
	}

	diff := vault.DiffSecrets(secrets, existing)
	if !diff.HasChanges() {
		log.Println("no changes detected, skipping write")
		return nil
	}

	log.Printf("changes: +%d -%d ~%d", len(diff.Added), len(diff.Removed), len(diff.Changed))

	if s.dryRun {
		log.Println("dry-run mode: skipping file write")
		return nil
	}

	if err := env.WriteEnvFile(s.output, secrets); err != nil {
		return fmt.Errorf("writing env file: %w", err)
	}

	return nil
}
