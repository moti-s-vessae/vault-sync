package sync

import (
	"context"
	"fmt"
)

// GroupStageOptions configures the GroupStage pipeline stage.
type GroupStageOptions struct {
	// By controls the grouping strategy (prefix or namespace).
	By GroupBy
	// Delimiter is the key separator used to determine group boundaries.
	Delimiter string
	// OnGroup is called for each group in sorted order. If nil the stage is a no-op.
	OnGroup func(group SecretGroup)
}

// GroupStage returns a pipeline Stage that partitions secrets into named groups
// and invokes OnGroup for each one. The original secret map is passed through
// unchanged so downstream stages continue to receive all secrets.
func GroupStage(opts GroupStageOptions) Stage {
	return func(ctx context.Context, secrets map[string]string) (map[string]string, error) {
		if opts.OnGroup == nil {
			return secrets, nil
		}

		grouper, err := NewSecretGrouper(opts.By, opts.Delimiter)
		if err != nil {
			return nil, fmt.Errorf("group stage: %w", err)
		}

		for _, grp := range grouper.Group(secrets) {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
			}
			opts.OnGroup(grp)
		}

		return secrets, nil
	}
}
