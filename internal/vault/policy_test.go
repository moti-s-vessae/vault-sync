package vault

import (
	"testing"
)

func basePolicy() *Policy {
	return &Policy{
		Rules: []PolicyRule{
			{Path: "secret/app/*", Capabilities: []string{"read", "list"}},
			{Path: "secret/shared/config", Capabilities: []string{"read"}},
			{Path: "secret/admin/*", Capabilities: []string{"*"}},
		},
	}
}

func TestCheckAccess_Allowed(t *testing.T) {
	p := basePolicy()
	if err := p.CheckAccess("secret/app/db", "read"); err != nil {
		t.Errorf("expected access, got: %v", err)
	}
}

func TestCheckAccess_WildcardCapability(t *testing.T) {
	p := basePolicy()
	if err := p.CheckAccess("secret/admin/token", "delete"); err != nil {
		t.Errorf("expected wildcard access, got: %v", err)
	}
}

func TestCheckAccess_MissingCapability(t *testing.T) {
	p := basePolicy()
	err := p.CheckAccess("secret/app/db", "write")
	if err == nil {
		t.Error("expected error for missing capability, got nil")
	}
}

func TestCheckAccess_NoMatchingRule(t *testing.T) {
	p := basePolicy()
	err := p.CheckAccess("secret/unknown/path", "read")
	if err == nil {
		t.Error("expected error for unmatched path, got nil")
	}
}

func TestCheckAccess_ExactPathMatch(t *testing.T) {
	p := basePolicy()
	if err := p.CheckAccess("secret/shared/config", "read"); err != nil {
		t.Errorf("expected exact path match, got: %v", err)
	}
}

func TestAllowedPaths_ReturnsMatchingPaths(t *testing.T) {
	p := basePolicy()
	paths := p.AllowedPaths("read")
	if len(paths) != 3 {
		t.Errorf("expected 3 readable paths, got %d", len(paths))
	}
}

func TestAllowedPaths_EmptyForUnknownCapability(t *testing.T) {
	p := basePolicy()
	paths := p.AllowedPaths("sudo")
	if len(paths) != 0 {
		t.Errorf("expected 0 paths for unknown capability, got %d", len(paths))
	}
}

func TestMatchesPolicy_WildcardPrefix(t *testing.T) {
	if !matchesPolicy("secret/app/*", "secret/app/db") {
		t.Error("expected wildcard match")
	}
}

func TestMatchesPolicy_NoMatch(t *testing.T) {
	if matchesPolicy("secret/app/*", "secret/other/db") {
		t.Error("expected no match")
	}
}
