package sync

import (
	"fmt"
	"log"

	"github.com/example/vault-sync/internal/config"
	"github.com/example/vault-sync/internal/env"
	"github.com/example/vault-sync/internal/vault"
)

// Result holds the outcome of a sync operation.
type Result struct {
	SecretsTotal   int
	SecretsWritten int
	OutputFile     string
}

// Syncer orchestrates fetching secrets from Vault and writing them to a .env file.
type Syncer struct {
	client *vault.Client
	cfg    *config.Config
}

// New creates a new Syncer from the provided config.
func New(cfg *config.Config) (*Syncer, error) {
	client, err := vault.NewClient(cfg.VaultAddr, cfg.VaultToken)
	if err != nil {
		return nil, fmt.Errorf("creating vault client: %w", err)
	}
	return &Syncer{client: client, cfg: cfg}, nil
}

// Run performs the full sync: fetch → filter → write.
func (s *Syncer) Run() (*Result, error) {
	log.Printf("fetching secrets from %s (path: %s)", s.cfg.VaultAddr, s.cfg.SecretPath)

	secrets, err := s.client.GetSecrets(s.cfg.SecretPath)
	if err != nil {
		return nil, fmt.Errorf("fetching secrets: %w", err)
	}

	total := len(secrets)

	if len(s.cfg.Prefixes) > 0 {
		secrets = vault.FilterSecrets(secrets, s.cfg.Prefixes)
		if s.cfg.StripPrefix {
			secrets = vault.StripPrefix(secrets, s.cfg.Prefixes)
		}
	}

	if err := env.WriteEnvFile(s.cfg.OutputFile, secrets); err != nil {
		return nil, fmt.Errorf("writing env file: %w", err)
	}

	log.Printf("wrote %d secret(s) to %s", len(secrets), s.cfg.OutputFile)

	return &Result{
		SecretsTotal:   total,
		SecretsWritten: len(secrets),
		OutputFile:     s.cfg.OutputFile,
	}, nil
}
