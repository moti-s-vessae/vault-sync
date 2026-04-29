package config

import (
	"errors"
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds all vault-sync configuration.
type Config struct {
	VaultAddr  string        `yaml:"vault_addr"`
	VaultToken string        `yaml:"vault_token"`
	SecretPath string        `yaml:"secret_path"`
	OutputFile string        `yaml:"output_file"`
	Prefixes   []string      `yaml:"prefixes"`
	Renames    []RenameRule  `yaml:"renames"`
	AuditLog   string        `yaml:"audit_log"`
	CachePath  string        `yaml:"cache_path"`
	CacheTTL   time.Duration `yaml:"cache_ttl"`
}

// RenameRule maps a vault key pattern to a new env var name.
type RenameRule struct {
	From string `yaml:"from"`
	To   string `yaml:"to"`
}

const defaultConfigFile = ".vault-sync.yaml"

// Load reads configuration from a YAML file and applies environment overrides.
func Load(path string) (*Config, error) {
	if path == "" {
		path = defaultConfigFile
	}

	cfg := &Config{
		VaultAddr:  "http://127.0.0.1:8200",
		OutputFile: ".env",
		CacheTTL:   5 * time.Minute,
	}

	data, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("reading config file: %w", err)
	}
	if err == nil {
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("parsing config file: %w", err)
		}
	}

	if v := os.Getenv("VAULT_ADDR"); v != "" {
		cfg.VaultAddr = v
	}
	if v := os.Getenv("VAULT_TOKEN"); v != "" {
		cfg.VaultToken = v
	}

	if cfg.VaultToken == "" {
		return nil, errors.New("vault token is required (set vault_token in config or VAULT_TOKEN env var)")
	}
	if cfg.SecretPath == "" {
		return nil, errors.New("secret_path is required")
	}

	return cfg, nil
}
