package vault

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// CacheEntry holds cached secrets with a timestamp.
type CacheEntry struct {
	Secrets   map[string]string `json:"secrets"`
	FetchedAt time.Time         `json:"fetched_at"`
}

// Cache manages on-disk caching of Vault secrets.
type Cache struct {
	path string
	ttl  time.Duration
}

// NewCache creates a Cache that stores data at path with the given TTL.
func NewCache(path string, ttl time.Duration) *Cache {
	return &Cache{path: path, ttl: ttl}
}

// Get returns cached secrets if they exist and are still valid.
// Returns nil, false if the cache is missing or expired.
func (c *Cache) Get() (map[string]string, bool) {
	data, err := os.ReadFile(c.path)
	if err != nil {
		return nil, false
	}
	var entry CacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, false
	}
	if time.Since(entry.FetchedAt) > c.ttl {
		return nil, false
	}
	return entry.Secrets, true
}

// Set writes secrets to the cache file.
func (c *Cache) Set(secrets map[string]string) error {
	if err := os.MkdirAll(filepath.Dir(c.path), 0700); err != nil {
		return err
	}
	entry := CacheEntry{
		Secrets:   secrets,
		FetchedAt: time.Now(),
	}
	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(c.path, data, 0600)
}

// Invalidate removes the cache file.
func (c *Cache) Invalidate() error {
	err := os.Remove(c.path)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}
