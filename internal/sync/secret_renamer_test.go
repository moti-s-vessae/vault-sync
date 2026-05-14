package sync

import (
	"testing"
)

func TestNewSecretRenamer_NoRules(t *testing.T) {
	_, err := NewSecretRenamer(nil)
	if err == nil {
		t.Fatal("expected error for empty rules")
	}
}

func TestNewSecretRenamer_EmptyPattern(t *testing.T) {
	_, err := NewSecretRenamer([]RenameRule{{Pattern: "", Replacement: "NEW"}})
	if err == nil {
		t.Fatal("expected error for empty pattern")
	}
}

func TestNewSecretRenamer_EmptyReplacement(t *testing.T) {
	_, err := NewSecretRenamer([]RenameRule{{Pattern: "FOO", Replacement: ""}})
	if err == nil {
		t.Fatal("expected error for empty replacement")
	}
}

func TestNewSecretRenamer_InvalidPattern(t *testing.T) {
	_, err := NewSecretRenamer([]RenameRule{{Pattern: "[", Replacement: "X"}})
	if err == nil {
		t.Fatal("expected error for invalid regex")
	}
}

func TestSecretRenamer_Apply_RenamesMatchingKey(t *testing.T) {
	r, err := NewSecretRenamer([]RenameRule{
		{Pattern: `^APP_(.+)$`, Replacement: `SVC_$1`},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out, err := r.Apply(map[string]string{"APP_TOKEN": "abc", "OTHER": "xyz"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out["SVC_TOKEN"] != "abc" {
		t.Errorf("expected SVC_TOKEN=abc, got %v", out)
	}
	if out["OTHER"] != "xyz" {
		t.Errorf("expected OTHER=xyz, got %v", out)
	}
}

func TestSecretRenamer_Apply_NoMatchKeepsOriginal(t *testing.T) {
	r, _ := NewSecretRenamer([]RenameRule{{Pattern: `^VAULT_`, Replacement: `V_`}})
	out, err := r.Apply(map[string]string{"DB_PASS": "secret"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out["DB_PASS"] != "secret" {
		t.Errorf("expected DB_PASS unchanged, got %v", out)
	}
}

func TestSecretRenamer_Apply_FirstMatchWins(t *testing.T) {
	r, _ := NewSecretRenamer([]RenameRule{
		{Pattern: `^FOO`, Replacement: `BAR`},
		{Pattern: `^FOO`, Replacement: `BAZ`},
	})
	out, err := r.Apply(map[string]string{"FOO_KEY": "v"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out["BAR_KEY"] != "v" {
		t.Errorf("expected BAR_KEY, got %v", out)
	}
}

func TestSecretRenamer_Apply_CollisionReturnsError(t *testing.T) {
	r, _ := NewSecretRenamer([]RenameRule{
		{Pattern: `^(A|B)$`, Replacement: `SAME`},
	})
	_, err := r.Apply(map[string]string{"A": "1", "B": "2"})
	if err == nil {
		t.Fatal("expected collision error")
	}
}

func TestRenamerStage_Integration_WithPipeline(t *testing.T) {
	loader := &mockLoader{secrets: map[string]string{"APP_DB": "host", "APP_KEY": "secret"}}
	p, err := NewPipeline(loader)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	r, err := NewSecretRenamer([]RenameRule{{Pattern: `^APP_`, Replacement: `SVC_`}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	p.AddStage(RenamerStage(r))
	out, err := p.Run(t.Context())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out["SVC_DB"] != "host" || out["SVC_KEY"] != "secret" {
		t.Errorf("unexpected output: %v", out)
	}
}
