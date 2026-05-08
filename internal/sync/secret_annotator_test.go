package sync

import (
	"testing"
)

func TestNewSecretAnnotator_NoRules(t *testing.T) {
	_, err := NewSecretAnnotator(nil)
	if err == nil {
		t.Fatal("expected error for empty rules")
	}
}

func TestNewSecretAnnotator_EmptyPattern(t *testing.T) {
	_, err := NewSecretAnnotator([]AnnotationRule{{Pattern: "", TagKey: "source", TagValue: "vault"}})
	if err == nil {
		t.Fatal("expected error for empty pattern")
	}
}

func TestNewSecretAnnotator_EmptyTagKey(t *testing.T) {
	_, err := NewSecretAnnotator([]AnnotationRule{{Pattern: ".*", TagKey: "", TagValue: "vault"}})
	if err == nil {
		t.Fatal("expected error for empty tag_key")
	}
}

func TestNewSecretAnnotator_InvalidPattern(t *testing.T) {
	_, err := NewSecretAnnotator([]AnnotationRule{{Pattern: "[", TagKey: "source", TagValue: "vault"}})
	if err == nil {
		t.Fatal("expected error for invalid regex")
	}
}

func TestSecretAnnotator_Annotate_MatchingKey(t *testing.T) {
	a, err := NewSecretAnnotator([]AnnotationRule{
		{Pattern: "^DB_", TagKey: "source", TagValue: "vault"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	input := map[string]string{"DB_PASSWORD": "secret", "APP_NAME": "myapp"}
	out := a.Annotate(input)

	if out["DB_PASSWORD"] != "secret" {
		t.Errorf("expected DB_PASSWORD to be preserved")
	}
	if out["DB_PASSWORD__source"] != "vault" {
		t.Errorf("expected annotation DB_PASSWORD__source=vault, got %q", out["DB_PASSWORD__source"])
	}
	if _, ok := out["APP_NAME__source"]; ok {
		t.Error("APP_NAME should not be annotated")
	}
}

func TestSecretAnnotator_Annotate_DoesNotMutateInput(t *testing.T) {
	a, _ := NewSecretAnnotator([]AnnotationRule{
		{Pattern: ".*", TagKey: "managed", TagValue: "true"},
	})
	input := map[string]string{"KEY": "val"}
	a.Annotate(input)
	if len(input) != 1 {
		t.Error("input map was mutated")
	}
}

func TestSecretAnnotator_Annotate_MultipleRules(t *testing.T) {
	a, err := NewSecretAnnotator([]AnnotationRule{
		{Pattern: "^DB_", TagKey: "tier", TagValue: "data"},
		{Pattern: "^DB_", TagKey: "managed", TagValue: "true"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := a.Annotate(map[string]string{"DB_HOST": "localhost"})
	if out["DB_HOST__tier"] != "data" {
		t.Errorf("expected tier annotation")
	}
	if out["DB_HOST__managed"] != "true" {
		t.Errorf("expected managed annotation")
	}
}

func TestAnnotateStage_PassesThroughSecrets(t *testing.T) {
	a, _ := NewSecretAnnotator([]AnnotationRule{
		{Pattern: "^API_", TagKey: "origin", TagValue: "vault"},
	})
	stage := AnnotateStage(a)
	input := map[string]string{"API_KEY": "abc123", "OTHER": "val"}
	out, err := stage(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out["API_KEY"] != "abc123" {
		t.Error("original secret not preserved")
	}
	if out["API_KEY__origin"] != "vault" {
		t.Error("annotation not injected")
	}
	if out["OTHER"] != "val" {
		t.Error("non-matching secret not preserved")
	}
}
