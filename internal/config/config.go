package config

import (
	"errors"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/user/vault-sync/internal/vault"
)

// Config holds all configuration for vault-sync.
type Config struct {
	VaultAddr  string   `yaml:"vault_addr"`
	VaultToken string   `yaml:"vault_token"`
	SecretPath string   `yaml:"secret_path"`
	OutputFile string   `yaml:"output_file"`
	Prefixes   []string `yaml:"prefixes"`
	AuditLog   string   `yaml:"audit_log"`

	Renames    []vault.RenameRule    `yaml:"renames"`
	Transforms []vault.TransformRule `yaml:"transforms"`
}

// Load reads configuration from a YAML file and overrides with environment variables.
func Load(path string) (*Config, error) {
	cfg := &Config{
		VaultAddr:  "http://127.0.0.1:8200",
		OutputFile: ".env",
	}

	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
		if err == nil {
			if err := yaml.Unmarshal(data, cfg); err != nil {
				return nil, err
			}
		}
	}

	if v := os.Getenv("VAULT_ADDR"); v != "" {
		cfg.VaultAddr = v
	}
	if v := os.Getenv("VAULT_TOKEN"); v != "" {
		cfg.VaultToken = v
	}
	if v := os.Getenv("VAULT_SECRET_PATH"); v != "" {
		cfg.SecretPath = v
	}
	if v := os.Getenv("VAULT_SYNC_OUTPUT"); v != "" {
		cfg.OutputFile = v
	}

	if cfg.VaultToken == "" {
		return nil, errors.New("vault token is required: set vault_token in config or VAULT_TOKEN env var")
	}
	return cfg, nil
}
