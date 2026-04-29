package env

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadEnvFile_BasicParsing(t *testing.T) {
	content := "FOO=bar\nBAZ=qux\n"
	path := writeTempEnv(t, content)

	result, err := ReadEnvFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["FOO"] != "bar" || result["BAZ"] != "qux" {
		t.Errorf("unexpected result: %v", result)
	}
}

func TestReadEnvFile_IgnoresComments(t *testing.T) {
	content := "# comment\nFOO=bar\n# another\n"
	path := writeTempEnv(t, content)

	result, err := ReadEnvFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 || result["FOO"] != "bar" {
		t.Errorf("unexpected result: %v", result)
	}
}

func TestReadEnvFile_UnquotesDoubleQuotes(t *testing.T) {
	content := `KEY="hello world"` + "\n"
	path := writeTempEnv(t, content)

	result, err := ReadEnvFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["KEY"] != "hello world" {
		t.Errorf("expected unquoted value, got %q", result["KEY"])
	}
}

func TestReadEnvFile_UnquotesSingleQuotes(t *testing.T) {
	content := "KEY='hello world'\n"
	path := writeTempEnv(t, content)

	result, err := ReadEnvFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["KEY"] != "hello world" {
		t.Errorf("expected unquoted value, got %q", result["KEY"])
	}
}

func TestReadEnvFile_NotExist(t *testing.T) {
	result, err := ReadEnvFile("/nonexistent/.env")
	if err != nil {
		t.Fatalf("expected no error for missing file, got: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty map, got %v", result)
	}
}

func writeTempEnv(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatalf("failed to write temp env: %v", err)
	}
	return path
}
