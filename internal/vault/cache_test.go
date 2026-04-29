package vault

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCache_SetAndGet_Valid(t *testing.T) {
	dir := t.TempDir()
	cachePath := filepath.Join(dir, "cache.json")
	c := NewCache(cachePath, 5*time.Minute)

	secrets := map[string]string{"KEY": "value", "FOO": "bar"}
	if err := c.Set(secrets); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	got, ok := c.Get()
	if !ok {
		t.Fatal("expected cache hit, got miss")
	}
	if got["KEY"] != "value" || got["FOO"] != "bar" {
		t.Errorf("unexpected secrets: %v", got)
	}
}

func TestCache_Get_Expired(t *testing.T) {
	dir := t.TempDir()
	cachePath := filepath.Join(dir, "cache.json")
	c := NewCache(cachePath, -1*time.Second) // already expired

	secrets := map[string]string{"KEY": "value"}
	if err := c.Set(secrets); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	_, ok := c.Get()
	if ok {
		t.Fatal("expected cache miss due to expiry, got hit")
	}
}

func TestCache_Get_Missing(t *testing.T) {
	dir := t.TempDir()
	c := NewCache(filepath.Join(dir, "nonexistent.json"), time.Minute)
	_, ok := c.Get()
	if ok {
		t.Fatal("expected cache miss for missing file")
	}
}

func TestCache_Invalidate(t *testing.T) {
	dir := t.TempDir()
	cachePath := filepath.Join(dir, "cache.json")
	c := NewCache(cachePath, time.Minute)

	if err := c.Set(map[string]string{"A": "1"}); err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	if err := c.Invalidate(); err != nil {
		t.Fatalf("Invalidate failed: %v", err)
	}
	if _, err := os.Stat(cachePath); !os.IsNotExist(err) {
		t.Error("expected cache file to be removed")
	}
}

func TestCache_Invalidate_NotExist(t *testing.T) {
	dir := t.TempDir()
	c := NewCache(filepath.Join(dir, "nope.json"), time.Minute)
	if err := c.Invalidate(); err != nil {
		t.Errorf("Invalidate on missing file should not error: %v", err)
	}
}

func TestCache_FilePermissions(t *testing.T) {
	dir := t.TempDir()
	cachePath := filepath.Join(dir, "cache.json")
	c := NewCache(cachePath, time.Minute)

	if err := c.Set(map[string]string{"SECRET": "s3cr3t"}); err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	info, err := os.Stat(cachePath)
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Errorf("expected permissions 0600, got %o", perm)
	}
}
