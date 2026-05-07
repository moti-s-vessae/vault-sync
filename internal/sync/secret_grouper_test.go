package sync

import (
	"testing"
)

func TestNewSecretGrouper_InvalidBy(t *testing.T) {
	_, err := NewSecretGrouper("unknown", "_")
	if err == nil {
		t.Fatal("expected error for unsupported GroupBy, got nil")
	}
}

func TestNewSecretGrouper_EmptyDelimiter(t *testing.T) {
	_, err := NewSecretGrouper(GroupByPrefix, "")
	if err == nil {
		t.Fatal("expected error for empty delimiter, got nil")
	}
}

func TestNewSecretGrouper_Valid(t *testing.T) {
	g, err := NewSecretGrouper(GroupByPrefix, "_")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g == nil {
		t.Fatal("expected non-nil grouper")
	}
}

func TestGroup_SplitsByFirstSegment(t *testing.T) {
	g, _ := NewSecretGrouper(GroupByPrefix, "_")
	secrets := map[string]string{
		"DB_HOST":    "localhost",
		"DB_PORT":    "5432",
		"APP_SECRET": "abc",
	}
	groups := g.Group(secrets)
	if len(groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(groups))
	}
	if groups[0].Name != "APP" {
		t.Errorf("expected first group APP, got %s", groups[0].Name)
	}
	if groups[1].Name != "DB" {
		t.Errorf("expected second group DB, got %s", groups[1].Name)
	}
	if len(groups[1].Secrets) != 2 {
		t.Errorf("expected 2 secrets in DB group, got %d", len(groups[1].Secrets))
	}
}

func TestGroup_NoDelimiterGoesToDefault(t *testing.T) {
	g, _ := NewSecretGrouper(GroupByPrefix, "_")
	secrets := map[string]string{
		"NODASH": "value",
	}
	groups := g.Group(secrets)
	if len(groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(groups))
	}
	if groups[0].Name != "default" {
		t.Errorf("expected group name 'default', got %s", groups[0].Name)
	}
}

func TestGroup_EmptySecrets(t *testing.T) {
	g, _ := NewSecretGrouper(GroupByNamespace, "/")
	groups := g.Group(map[string]string{})
	if len(groups) != 0 {
		t.Errorf("expected 0 groups for empty input, got %d", len(groups))
	}
}

func TestGroup_SortedGroupNames(t *testing.T) {
	g, _ := NewSecretGrouper(GroupByPrefix, "_")
	secrets := map[string]string{
		"Z_KEY": "1",
		"A_KEY": "2",
		"M_KEY": "3",
	}
	groups := g.Group(secrets)
	expected := []string{"A", "M", "Z"}
	for i, grp := range groups {
		if grp.Name != expected[i] {
			t.Errorf("index %d: expected %s, got %s", i, expected[i], grp.Name)
		}
	}
}
