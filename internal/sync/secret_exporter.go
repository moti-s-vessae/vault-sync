package sync

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

// ExportFormat defines the output format for exported secrets.
type ExportFormat string

const (
	FormatEnv  ExportFormat = "env"
	FormatJSON ExportFormat = "json"
	FormatDotenv ExportFormat = "dotenv"
)

// SecretExporter writes secrets to an io.Writer in a specified format.
type SecretExporter struct {
	format ExportFormat
	out    io.Writer
}

// NewSecretExporter creates a SecretExporter for the given format.
// If out is nil, os.Stdout is used.
func NewSecretExporter(format ExportFormat, out io.Writer) (*SecretExporter, error) {
	switch format {
	case FormatEnv, FormatJSON, FormatDotenv:
	default:
		return nil, fmt.Errorf("unsupported export format: %q", format)
	}
	if out == nil {
		out = os.Stdout
	}
	return &SecretExporter{format: format, out: out}, nil
}

// Export writes secrets to the configured writer in the configured format.
func (e *SecretExporter) Export(secrets map[string]string) error {
	switch e.format {
	case FormatJSON:
		return e.exportJSON(secrets)
	case FormatEnv, FormatDotenv:
		return e.exportEnv(secrets)
	default:
		return fmt.Errorf("unsupported format: %q", e.format)
	}
}

func (e *SecretExporter) exportJSON(secrets map[string]string) error {
	enc := json.NewEncoder(e.out)
	enc.SetIndent("", "  ")
	return enc.Encode(secrets)
}

func (e *SecretExporter) exportEnv(secrets map[string]string) error {
	keys := make([]string, 0, len(secrets))
	for k := range secrets {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder
	for _, k := range keys {
		v := secrets[k]
		if strings.ContainsAny(v, " \t\n#") {
			v = fmt.Sprintf(`"%s"`, strings.ReplaceAll(v, `"`, `\"`))
		}
		sb.WriteString(fmt.Sprintf("%s=%s\n", k, v))
	}
	_, err := io.WriteString(e.out, sb.String())
	return err
}
