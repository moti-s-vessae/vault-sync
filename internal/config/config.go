package config

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds all runtime configuration for vault-sync.
type Config struct {
	VaultAddr   string   `yaml:"vault_addr"`
	VaultToken  string   `yaml:"vault_token"`
	SecretPath  string   `yaml:"secret_path"`
	OutputFile  string   `yaml:"output_file"`
	Prefixes    []string `yaml:"prefixes"`
	StripPrefix bool     `yaml:"strip_prefix"`
}

// Load reads config from file (if it exists) then overlays environment variables.
func Load(cfgFile string) (*Config, error) {
	cfg := &Config{
		VaultAddr:  "http://127.0.0.1:8200",
		OutputFile: ".env",
	}

	if data, err := os.ReadFile(cfgFile); err == nil {
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("parsing config file %q: %w", cfgFile, err)
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
		return nil, errors.New("vault token is required (set VAULT_TOKEN or vault_token in config)")
	}
	if cfg.SecretPath == "" {
		return nil, errors.New("secret_path is required")
	}

	return cfg, nil
}
