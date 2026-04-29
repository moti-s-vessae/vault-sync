package env

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteEnvFile_BasicSecrets(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")

	secrets := map[string]string{
		"DB_HOST":     "localhost",
		"DB_PORT":     "5432",
		"DB_PASSWORD": "s3cr3t",
	}

	if err := WriteEnvFile(path, secrets); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read written file: %v", err)
	}

	for _, expected := range []string{"DB_HOST=localhost", "DB_PORT=5432", "DB_PASSWORD=s3cr3t"} {
		if !strings.Contains(string(content), expected) {
			t.Errorf("expected %q in output, got:\n%s", expected, content)
		}
	}
}

func TestWriteEnvFile_QuotesSpecialValues(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")

	secrets := map[string]string{
		"MSG": "hello world",
		"TOKEN": `abc"def`,
	}

	if err := WriteEnvFile(path, secrets); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content := string(must(os.ReadFile(path)))

	if !strings.Contains(content, `MSG="hello world"`) {
		t.Errorf("expected quoted MSG, got:\n%s", content)
	}
	if !strings.Contains(content, `TOKEN="abc\"def"`) {
		t.Errorf("expected escaped TOKEN, got:\n%s", content)
	}
}

func TestWriteEnvFile_SortedOutput(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")

	secrets := map[string]string{
		"Z_KEY": "z",
		"A_KEY": "a",
		"M_KEY": "m",
	}

	if err := WriteEnvFile(path, secrets); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content := string(must(os.ReadFile(path)))
	lines := strings.Split(strings.TrimSpace(content), "\n")

	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}
	if !strings.HasPrefix(lines[0], "A_KEY") || !strings.HasPrefix(lines[1], "M_KEY") || !strings.HasPrefix(lines[2], "Z_KEY") {
		t.Errorf("expected sorted output, got: %v", lines)
	}
}

func TestWriteEnvFile_FilePermissions(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")

	if err := WriteEnvFile(path, map[string]string{"KEY": "val"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("failed to stat file: %v", err)
	}
	if info.Mode().Perm() != 0600 {
		t.Errorf("expected permissions 0600, got %v", info.Mode().Perm())
	}
}

func must(b []byte, err error) []byte {
	if err != nil {
		panic(err)
	}
	return b
}
