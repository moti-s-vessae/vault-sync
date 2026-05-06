package sync

import (
	"strings"
	"testing"
)

func TestNewSecretTemplater_EmptyTemplate(t *testing.T) {
	_, err := NewSecretTemplater("")
	if err == nil {
		t.Fatal("expected error for empty template")
	}
}

func TestNewSecretTemplater_InvalidTemplate(t *testing.T) {
	_, err := NewSecretTemplater("{{ .Foo")
	if err == nil {
		t.Fatal("expected error for invalid template")
	}
}

func TestNewSecretTemplater_Valid(t *testing.T) {
	st, err := NewSecretTemplater("hello {{ index . \"NAME\" }}")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if st == nil {
		t.Fatal("expected non-nil SecretTemplater")
	}
}

func TestSecretTemplater_Render_Basic(t *testing.T) {
	st, _ := NewSecretTemplater(`postgres://{{ index . "DB_USER" }}:{{ index . "DB_PASS" }}@localhost/mydb`)
	result, err := st.Render(map[string]string{
		"DB_USER": "admin",
		"DB_PASS": "s3cr3t",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "postgres://admin:s3cr3t@localhost/mydb"
	if result != want {
		t.Errorf("got %q, want %q", result, want)
	}
}

func TestSecretTemplater_Render_MissingKey(t *testing.T) {
	st, _ := NewSecretTemplater(`{{ index . "MISSING_KEY" }}`)
	_, err := st.Render(map[string]string{"OTHER": "val"})
	if err == nil {
		t.Fatal("expected error for missing key")
	}
	if !strings.Contains(err.Error(), "render failed") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestSecretTemplater_Render_NilSecrets(t *testing.T) {
	st, _ := NewSecretTemplater(`static-value`)
	result, err := st.Render(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "static-value" {
		t.Errorf("got %q, want %q", result, "static-value")
	}
}

func TestTemplateStage_AddsOutputKey(t *testing.T) {
	stage := TemplateStage(`{{ index . "HOST" }}:{{ index . "PORT" }}`, "DSN")
	input := map[string]string{"HOST": "localhost", "PORT": "5432"}
	out, err := stage(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out["DSN"] != "localhost:5432" {
		t.Errorf("got DSN=%q, want %q", out["DSN"], "localhost:5432")
	}
	if out["HOST"] != "localhost" {
		t.Error("original secrets should be preserved")
	}
}

func TestTemplateStage_InvalidTemplate_ReturnsError(t *testing.T) {
	stage := TemplateStage("", "OUT")
	_, err := stage(map[string]string{})
	if err == nil {
		t.Fatal("expected error for invalid template")
	}
}

func TestTemplateStage_DoesNotMutateInput(t *testing.T) {
	stage := TemplateStage(`hello`, "GREETING")
	input := map[string]string{"A": "1"}
	_, err := stage(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := input["GREETING"]; ok {
		t.Error("stage must not mutate the input map")
	}
}
