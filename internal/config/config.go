package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

// Config holds the application configuration.
type Config struct {
	VaultAddr  string `mapstructure:"vault_addr"`
	VaultToken string `mapstructure:"vault_token"`
	Namespace  string `mapstructure:"namespace"`
	OutputFile string `mapstructure:"output_file"`
	MountPath  string `mapstructure:"mount_path"`
}

// Load reads configuration from a file and environment variables.
// Environment variables take precedence over file values.
func Load(cfgFile string) (*Config, error) {
	v := viper.New()

	v.SetDefault("vault_addr", "http://127.0.0.1:8200")
	v.SetDefault("output_file", ".env")
	v.SetDefault("mount_path", "secret")

	v.SetEnvPrefix("VAULT_SYNC")
	v.AutomaticEnv()

	// Also bind standard Vault env vars.
	_ = v.BindEnv("vault_addr", "VAULT_ADDR")
	_ = v.BindEnv("vault_token", "VAULT_TOKEN")

	if cfgFile != "" {
		v.SetConfigFile(cfgFile)
	} else {
		v.AddConfigPath(".")
		v.SetConfigName(".vault-sync")
		v.SetConfigType("yaml")
	}

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(*os.PathError); !ok && cfgFile != "" {
			return nil, fmt.Errorf("reading config file: %w", err)
		}
		// Config file is optional; continue with defaults + env vars.
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshalling config: %w", err)
	}

	if cfg.VaultToken == "" {
		return nil, fmt.Errorf("vault token is required (set VAULT_TOKEN or vault_token in config)")
	}

	return &cfg, nil
}
