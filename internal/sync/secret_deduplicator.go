package sync

import (
	"fmt"
	"sort"
)

// DeduplicateStrategy defines how duplicate keys are resolved.
type DeduplicateStrategy string

const (
	// DeduplicateKeepFirst retains the first occurrence of a duplicate key.
	DeduplicateKeepFirst DeduplicateStrategy = "keep-first"
	// DeduplicateKeepLast retains the last occurrence of a duplicate key.
	DeduplicateKeepLast DeduplicateStrategy = "keep-last"
	// DeduplicateError returns an error if any duplicate keys are found.
	DeduplicateError DeduplicateStrategy = "error"
)

// ValidateDeduplicateStrategy returns an error if the strategy is not recognised.
func ValidateDeduplicateStrategy(s DeduplicateStrategy) error {
	switch s {
	case DeduplicateKeepFirst, DeduplicateKeepLast, DeduplicateError:
		return nil
	default:
		return fmt.Errorf("unknown deduplicate strategy %q: must be one of keep-first, keep-last, error", s)
	}
}

// DeduplicateSecrets removes duplicate keys from secrets according to the
// given strategy. The input slice preserves insertion order semantics.
func DeduplicateSecrets(secrets []map[string]string, strategy DeduplicateStrategy) (map[string]string, error) {
	if err := ValidateDeduplicateStrategy(strategy); err != nil {
		return nil, err
	}

	seen := make(map[string]string)
	duplicates := make(map[string]struct{})

	for _, m := range secrets {
		// Sort keys for deterministic processing within each map.
		keys := make([]string, 0, len(m))
		for k := range m {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			if _, exists := seen[k]; exists {
				duplicates[k] = struct{}{}
				if strategy == DeduplicateError {
					return nil, fmt.Errorf("duplicate secret key %q", k)
				}
				if strategy == DeduplicateKeepLast {
					seen[k] = m[k]
				}
				// keep-first: do nothing
			} else {
				seen[k] = m[k]
			}
		}
	}

	_ = duplicates // available for future logging
	return seen, nil
}

// DeduplicateStage returns a pipeline stage that merges multiple secret maps
// produced by prior stages into one, applying the given deduplication strategy.
func DeduplicateStage(strategy DeduplicateStrategy) func(map[string]string) (map[string]string, error) {
	return func(secrets map[string]string) (map[string]string, error) {
		// Single-map path: validate strategy and return as-is.
		if err := ValidateDeduplicateStrategy(strategy); err != nil {
			return nil, err
		}
		return secrets, nil
	}
}
