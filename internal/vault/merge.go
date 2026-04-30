package vault

// MergeStrategy defines how conflicts between existing and new secrets are resolved.
type MergeStrategy int

const (
	// MergeStrategyOverwrite replaces existing values with new ones.
	MergeStrategyOverwrite MergeStrategy = iota
	// MergeStrategyKeepExisting preserves existing values when a key already exists.
	MergeStrategyKeepExisting
	// MergeStrategyError returns an error if a conflict is detected.
	MergeStrategyError
)

// MergeConflict describes a key that exists in both base and incoming maps with different values.
type MergeConflict struct {
	Key      string
	BaseVal  string
	NewVal   string
}

// MergeResult holds the merged secrets and any conflicts that were encountered.
type MergeResult struct {
	Secrets   map[string]string
	Conflicts []MergeConflict
}

// MergeSecrets merges incoming secrets into base according to the given strategy.
// base represents the current local state; incoming represents the remote Vault state.
func MergeSecrets(base, incoming map[string]string, strategy MergeStrategy) (MergeResult, error) {
	result := MergeResult{
		Secrets: make(map[string]string, len(base)),
	}

	// Seed result with base values.
	for k, v := range base {
		result.Secrets[k] = v
	}

	for k, inVal := range incoming {
		baseVal, exists := base[k]
		if !exists {
			// New key — always add.
			result.Secrets[k] = inVal
			continue
		}

		if baseVal == inVal {
			// No conflict.
			continue
		}

		conflict := MergeConflict{Key: k, BaseVal: baseVal, NewVal: inVal}
		result.Conflicts = append(result.Conflicts, conflict)

		switch strategy {
		case MergeStrategyOverwrite:
			result.Secrets[k] = inVal
		case MergeStrategyKeepExisting:
			// Keep base value — already set.
		case MergeStrategyError:
			return MergeResult{}, &MergeError{Conflicts: result.Conflicts}
		}
	}

	return result, nil
}

// MergeError is returned when MergeStrategyError is used and conflicts are found.
type MergeError struct {
	Conflicts []MergeConflict
}

func (e *MergeError) Error() string {
	return "merge conflict: one or more keys differ between local and remote state"
}
