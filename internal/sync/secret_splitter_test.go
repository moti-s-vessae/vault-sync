package sync

import (
	"testing"
)

func TestNewSecretSplitter_NoRules(t *testing.T) {
	_, err := NewSecretSplitter(nil)
	if err == nil {
		t.Fatal("expected error for nil rules")
	}
}

func TestNewSecretSplitter_EmptyPattern(t *testing.T) {
	_, err := NewSecretSplitter([]SplitRule{{Pattern: "", Separator: ","}})
	if err == nil {
		t.Fatal("expected error for empty pattern")
	}
}

func TestNewSecretSplitter_EmptySeparator(t *testing.T) {
	_, err := NewSecretSplitter([]SplitRule{{Pattern: "*", Separator: ""}})
	if err == nil {
		t.Fatal("expected error for empty separator")
	}
}

func TestNewSecretSplitter_Valid(t *testing.T) {
	s, err := NewSecretSplitter([]SplitRule{{Pattern: "HOSTS", Separator: ","}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s == nil {
		t.Fatal("expected non-nil splitter")
	}
}

func TestSecretSplitter_Apply_SplitsMatchingKey(t *testing.T) {
	s, _ := NewSecretSplitter([]SplitRule{
		{Pattern: "HOSTS", Separator: ",", KeyTemplate: "{{.Key}}_{{.Index}}"},
	})
	secrets := map[string]string{"HOSTS": "a,b,c", "OTHER": "x"}
	out, err := s.Apply(secrets)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := out["HOSTS"]; ok {
		t.Error("original key should be replaced")
	}
	for i, want := range []string{"a", "b", "c"} {
		key := "HOSTS_" + string(rune('1'+i))
		if out[key] != want {
			t.Errorf("out[%q] = %q, want %q", key, out[key], want)
		}
	}
	if out["OTHER"] != "x" {
		t.Errorf("non-matching key should pass through")
	}
}

func TestSecretSplitter_Apply_NoMatchPassesThrough(t *testing.T) {
	s, _ := NewSecretSplitter([]SplitRule{{Pattern: "HOSTS", Separator: ","}})
	secrets := map[string]string{"DB_URL": "postgres://localhost"}
	out, err := s.Apply(secrets)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out["DB_URL"] != "postgres://localhost" {
		t.Errorf("expected pass-through for non-matching key")
	}
}

func TestSecretSplitter_Apply_DefaultKeyTemplate(t *testing.T) {
	s, _ := NewSecretSplitter([]SplitRule{{Pattern: "TAGS", Separator: ";"}}) // no KeyTemplate
	out, err := s.Apply(map[string]string{"TAGS": "foo;bar"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out["TAGS_1"] != "foo" || out["TAGS_2"] != "bar" {
		t.Errorf("unexpected output: %v", out)
	}
}

func TestSecretSplitter_Apply_WildcardPattern(t *testing.T) {
	s, _ := NewSecretSplitter([]SplitRule{{Pattern: "APP_*", Separator: "|"}})
	secrets := map[string]string{"APP_HOSTS": "h1|h2", "DB": "val"}
	out, err := s.Apply(secrets)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out["APP_HOSTS_1"] != "h1" || out["APP_HOSTS_2"] != "h2" {
		t.Errorf("wildcard split failed: %v", out)
	}
	if out["DB"] != "val" {
		t.Error("non-matching key should pass through")
	}
}
