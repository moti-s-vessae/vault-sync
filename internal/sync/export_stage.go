package sync

import (
	"fmt"
	"io"
)

// ExportStageOptions configures the export pipeline stage.
type ExportStageOptions struct {
	Format ExportFormat
	Out    io.Writer
}

// ExportStage returns a pipeline Stage that exports secrets using SecretExporter.
// It does not modify the secrets map; export is a side-effect only.
func ExportStage(opts ExportStageOptions) Stage {
	return func(secrets map[string]string) (map[string]string, error) {
		if opts.Format == "" {
			return secrets, nil
		}
		exporter, err := NewSecretExporter(opts.Format, opts.Out)
		if err != nil {
			return nil, fmt.Errorf("export stage: %w", err)
		}
		if err := exporter.Export(secrets); err != nil {
			return nil, fmt.Errorf("export stage: write failed: %w", err)
		}
		return secrets, nil
	}
}
