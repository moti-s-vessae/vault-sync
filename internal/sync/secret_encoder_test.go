package sync

import (
	"encoding/base64"
	"testing"
)

func TestNewSecretEncoder_NoRules(t *testing.T) {
	_, err := NewSecretEncoder(nil)
	if err == nil {
		t.Fatal("expected error for empty rules")
	}
}

func TestNewSecretEncoder_InvalidPattern(t *testing.T) {
	_, err := NewSecretEncoder([]EncodeRule{
		{Pattern: "[invalid", Format: "base64"},
	})
	if err == nil {
		t.Fatal("expected error for invalid regex pattern")
	}
}

func TestNewSecretEncoder_UnsupportedFormat(t *testing.T) {
	_, err := NewSecretEncoder([]EncodeRule{
		{Pattern: ".*", Format: "hex"},
	})
	if err == nil {
		t.Fatal("expected error for unsupported format")
	}
}

func TestSecretEncoder_Encode_Base64(t *testing.T) {
	enc, err := NewSecretEncoder([]EncodeRule{
		{Pattern: "^SECRET_", Format: "base64"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	secrets := map[string]string{
		"SECRET_KEY": "mysecret",
		"PLAIN_KEY":  "plainvalue",
	}
	out := enc.Encode(secrets)

	want := base64.StdEncoding.EncodeToString([]byte("mysecret"))
	if out["SECRET_KEY"] != want {
		t.Errorf("SECRET_KEY: got %q, want %q", out["SECRET_KEY"], want)
	}
	if out["PLAIN_KEY"] != "plainvalue" {
		t.Errorf("PLAIN_KEY should be unchanged, got %q", out["PLAIN_KEY"])
	}
}

func TestSecretEncoder_Encode_Base64URL(t *testing.T) {
	enc, err := NewSecretEncoder([]EncodeRule{
		{Pattern: "_TOKEN$", Format: "base64url"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	secrets := map[string]string{"API_TOKEN": "abc+/def"}
	out := enc.Encode(secrets)

	want := base64.URLEncoding.EncodeToString([]byte("abc+/def"))
	if out["API_TOKEN"] != want {
		t.Errorf("API_TOKEN: got %q, want %q", out["API_TOKEN"], want)
	}
}

func TestSecretEncoder_DoesNotMutateInput(t *testing.T) {
	enc, _ := NewSecretEncoder([]EncodeRule{
		{Pattern: ".*", Format: "base64"},
	})
	input := map[string]string{"KEY": "value"}
	_ = enc.Encode(input)
	if input["KEY"] != "value" {
		t.Error("Encode mutated the input map")
	}
}

func TestEncodeStage_Integration_WithPipeline(t *testing.T) {
	enc, err := NewSecretEncoder([]EncodeRule{
		{Pattern: "^ENC_", Format: "base64"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	loader := &mockLoader{
		secrets: map[string]string{
			"ENC_PASS": "hunter2",
			"DB_HOST":  "localhost",
		},
	}

	p, err := NewPipeline(loader, EncodeStage(enc))
	if err != nil {
		t.Fatalf("pipeline error: %v", err)
	}

	out, err := p.Run(t.Context(), "secret/data/app")
	if err != nil {
		t.Fatalf("run error: %v", err)
	}

	want := base64.StdEncoding.EncodeToString([]byte("hunter2"))
	if out["ENC_PASS"] != want {
		t.Errorf("ENC_PASS: got %q, want %q", out["ENC_PASS"], want)
	}
	if out["DB_HOST"] != "localhost" {
		t.Errorf("DB_HOST should be unchanged")
	}
}
