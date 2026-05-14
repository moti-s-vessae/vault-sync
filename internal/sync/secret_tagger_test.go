package sync

import (
	"testing"
)

func TestNewSecretTagger_NoRules(t *testing.T) {
	_, err := NewSecretTagger(nil)
	if err == nil {
		t.Fatal("expected error for nil rules")
	}
}

func TestNewSecretTagger_EmptyPattern(t *testing.T) {
	_, err := NewSecretTagger([]TagRule{{Pattern: "", Tag: "env"}})
	if err == nil {
		t.Fatal("expected error for empty pattern")
	}
}

func TestNewSecretTagger_EmptyTag(t *testing.T) {
	_, err := NewSecretTagger([]TagRule{{Pattern: ".*", Tag: ""}})
	if err == nil {
		t.Fatal("expected error for empty tag")
	}
}

func TestNewSecretTagger_InvalidPattern(t *testing.T) {
	_, err := NewSecretTagger([]TagRule{{Pattern: "[", Tag: "env"}})
	if err == nil {
		t.Fatal("expected error for invalid regex")
	}
}

func TestNewSecretTagger_Valid(t *testing.T) {
	tagger, err := NewSecretTagger([]TagRule{{Pattern: "^DB_", Tag: "db"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tagger == nil {
		t.Fatal("expected non-nil tagger")
	}
}

func TestSecretTagger_Apply_MatchingKey(t *testing.T) {
	tagger, _ := NewSecretTagger([]TagRule{{Pattern: "^DB_", Tag: "db"}})
	input := map[string]string{"DB_HOST": "localhost", "APP_PORT": "8080"}
	out, err := tagger.Apply(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := out["db:DB_HOST"]; !ok {
		t.Error("expected db:DB_HOST in output")
	}
	if _, ok := out["APP_PORT"]; !ok {
		t.Error("expected APP_PORT unchanged in output")
	}
	if _, ok := out["DB_HOST"]; ok {
		t.Error("original DB_HOST should not appear in output")
	}
}

func TestSecretTagger_Apply_NoMatchKeepsOriginal(t *testing.T) {
	tagger, _ := NewSecretTagger([]TagRule{{Pattern: "^SECRET_", Tag: "sec"}})
	input := map[string]string{"APP_KEY": "value"}
	out, err := tagger.Apply(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v, ok := out["APP_KEY"]; !ok || v != "value" {
		t.Errorf("expected APP_KEY=value, got %v", out)
	}
}

func TestSecretTagger_Apply_DoesNotMutateInput(t *testing.T) {
	tagger, _ := NewSecretTagger([]TagRule{{Pattern: ".*", Tag: "all"}})
	input := map[string]string{"KEY": "val"}
	_, err := tagger.Apply(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := input["all:KEY"]; ok {
		t.Error("Apply must not mutate the input map")
	}
}

func TestSecretTagger_Apply_CollisionReturnsError(t *testing.T) {
	tagger, _ := NewSecretTagger([]TagRule{
		{Pattern: "^A", Tag: "t"},
		{Pattern: "^B", Tag: "t"},
	})
	// Manually craft a scenario where two keys collide after tagging.
	// "t:AKEY" from AKEY, and a pre-existing key named "t:AKEY" would collide.
	input := map[string]string{"AKEY": "1", "t:AKEY": "2"}
	_, err := tagger.Apply(input)
	if err == nil {
		t.Fatal("expected collision error")
	}
}
