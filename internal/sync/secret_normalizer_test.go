package sync

import (
	"testing"
)

func TestNewSecretNormalizer_NoRules(t *testing.T) {
	_, err := NewSecretNormalizer(nil)
	if err == nil {
		t.Fatal("expected error for empty rules")
	}
}

func TestNewSecretNormalizer_EmptyPattern(t *testing.T) {
	_, err := NewSecretNormalizer([]NormalizeRule{{Pattern: "", Strategy: "upper"}})
	if err == nil {
		t.Fatal("expected error for empty pattern")
	}
}

func TestNewSecretNormalizer_UnsupportedStrategy(t *testing.T) {
	_, err := NewSecretNormalizer([]NormalizeRule{{Pattern: ".*", Strategy: "title"}})
	if err == nil {
		t.Fatal("expected error for unsupported strategy")
	}
}

func TestNewSecretNormalizer_InvalidPattern(t *testing.T) {
	_, err := NewSecretNormalizer([]NormalizeRule{{Pattern: "[invalid", Strategy: "upper"}})
	if err == nil {
		t.Fatal("expected error for invalid regex")
	}
}

func TestSecretNormalizer_Apply_Upper(t *testing.T) {
	n, err := NewSecretNormalizer([]NormalizeRule{{Pattern: ".*", Strategy: "upper"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out, err := n.Apply(map[string]string{"db_host": "localhost", "api_key": "secret"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out["DB_HOST"] != "localhost" {
		t.Errorf("expected DB_HOST=localhost, got %v", out)
	}
	if out["API_KEY"] != "secret" {
		t.Errorf("expected API_KEY=secret, got %v", out)
	}
}

func TestSecretNormalizer_Apply_Lower(t *testing.T) {
	n, _ := NewSecretNormalizer([]NormalizeRule{{Pattern: ".*", Strategy: "lower"}})
	out, err := n.Apply(map[string]string{"DB_HOST": "localhost"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out["db_host"] != "localhost" {
		t.Errorf("expected db_host, got %v", out)
	}
}

func TestSecretNormalizer_Apply_Snake(t *testing.T) {
	n, _ := NewSecretNormalizer([]NormalizeRule{{Pattern: ".*", Strategy: "snake"}})
	out, err := n.Apply(map[string]string{"db-host": "localhost"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out["DB_HOST"] != "localhost" {
		t.Errorf("expected DB_HOST, got %v", out)
	}
}

func TestSecretNormalizer_Apply_Camel(t *testing.T) {
	n, _ := NewSecretNormalizer([]NormalizeRule{{Pattern: ".*", Strategy: "camel"}})
	out, err := n.Apply(map[string]string{"db_host": "localhost"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out["dbHost"] != "localhost" {
		t.Errorf("expected dbHost, got %v", out)
	}
}

func TestSecretNormalizer_Apply_NoMatchKeepsOriginal(t *testing.T) {
	n, _ := NewSecretNormalizer([]NormalizeRule{{Pattern: "^prefix_", Strategy: "upper"}})
	out, err := n.Apply(map[string]string{"other_key": "value"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out["other_key"] != "value" {
		t.Errorf("expected key unchanged, got %v", out)
	}
}

func TestSecretNormalizer_Apply_CollisionReturnsError(t *testing.T) {
	n, _ := NewSecretNormalizer([]NormalizeRule{{Pattern: ".*", Strategy: "upper"}})
	_, err := n.Apply(map[string]string{"db_host": "a", "DB_HOST": "b"})
	if err == nil {
		t.Fatal("expected collision error")
	}
}

func TestNormalizeStage_Integration(t *testing.T) {
	n, _ := NewSecretNormalizer([]NormalizeRule{{Pattern: ".*", Strategy: "upper"}})
	stage := NormalizeStage(n)
	out, err := stage(map[string]string{"foo": "bar"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out["FOO"] != "bar" {
		t.Errorf("expected FOO=bar, got %v", out)
	}
}
