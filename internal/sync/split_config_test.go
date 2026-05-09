package sync

import (
	"testing"
)

func TestSplitConfig_Validate_Disabled(t *testing.T) {
	c := &SplitConfig{Enabled: false}
	if err := c.Validate(); err != nil {
		t.Fatalf("expected no error for disabled config, got %v", err)
	}
}

func TestSplitConfig_Validate_EnabledNoRules(t *testing.T) {
	c := &SplitConfig{Enabled: true, Rules: nil}
	if err := c.Validate(); err == nil {
		t.Fatal("expected error for enabled config with no rules")
	}
}

func TestSplitConfig_Validate_EmptyPattern(t *testing.T) {
	c := &SplitConfig{
		Enabled: true,
		Rules:   []SplitRule{{Pattern: "", Separator: ","}},
	}
	if err := c.Validate(); err == nil {
		t.Fatal("expected error for empty pattern")
	}
}

func TestSplitConfig_Validate_EmptySeparator(t *testing.T) {
	c := &SplitConfig{
		Enabled: true,
		Rules:   []SplitRule{{Pattern: "HOSTS", Separator: ""}},
	}
	if err := c.Validate(); err == nil {
		t.Fatal("expected error for empty separator")
	}
}

func TestSplitConfig_Validate_Valid(t *testing.T) {
	c := &SplitConfig{
		Enabled: true,
		Rules:   []SplitRule{{Pattern: "HOSTS", Separator: ","}},
	}
	if err := c.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSplitConfig_ToSplitter_Disabled(t *testing.T) {
	c := &SplitConfig{Enabled: false}
	s, err := c.ToSplitter()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s != nil {
		t.Error("expected nil splitter for disabled config")
	}
}

func TestSplitConfig_ToSplitter_Valid(t *testing.T) {
	c := &SplitConfig{
		Enabled: true,
		Rules:   []SplitRule{{Pattern: "IPS", Separator: ";"}},
	}
	s, err := c.ToSplitter()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s == nil {
		t.Fatal("expected non-nil splitter")
	}
}
