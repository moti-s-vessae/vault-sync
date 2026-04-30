package vault

import (
	"fmt"
	"time"
)

// LoadOptions configures how secrets are loaded from Vault.
type LoadOptions struct {
	Prefixes   []string
	Renames    []RenameRule
	Transforms []TransformRule
	CacheTTL   time.Duration
}

// SecretsLoader orchestrates fetching, filtering, renaming, and transforming secrets.
type SecretsLoader struct {
	client SecretGetter
	cache  *Cache
}

// SecretGetter is the interface for fetching raw secrets from Vault.
type SecretGetter interface {
	GetSecrets(path string) (map[string]string, error)
}

// NewSecretsLoader creates a new SecretsLoader with the given client and cache.
func NewSecretsLoader(client SecretGetter, cache *Cache) *SecretsLoader {
	return &SecretsLoader{
		client: client,
		cache:  cache,
	}
}

// Load fetches secrets from the given Vault path and applies filtering,
// renaming, and transformation according to opts.
func (l *SecretsLoader) Load(path string, opts LoadOptions) (map[string]string, error) {
	if path == "" {
		return nil, fmt.Errorf("vault path must not be empty")
	}

	var raw map[string]string

	if l.cache != nil {
		if cached, ok := l.cache.Get(path); ok {
			raw = cached
		}
	}

	if raw == nil {
		var err error
		raw, err = l.client.GetSecrets(path)
		if err != nil {
			return nil, fmt.Errorf("loading secrets from %q: %w", path, err)
		}
		if l.cache != nil && opts.CacheTTL > 0 {
			l.cache.Set(path, raw, opts.CacheTTL)
		}
	}

	filtered := FilterSecrets(raw, opts.Prefixes)
	stripped := StripPrefix(filtered, opts.Prefixes)
	renamed := ApplyRenames(stripped, opts.Renames)
	transformed, err := TransformSecrets(renamed, opts.TransformRules())
	if err != nil {
		return nil, fmt.Errorf("transforming secrets: %w", err)
	}

	return transformed, nil
}

// TransformRules is a helper on LoadOptions to return the transform rules slice.
func (o LoadOptions) TransformRules() []TransformRule {
	return o.Transforms
}
