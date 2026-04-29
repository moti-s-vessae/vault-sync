package vault

// SecretChange describes a single change between two secret maps.
type SecretChange struct {
	Key    string
	Action string // "added", "removed", "changed", "unchanged"
	OldVal string
	NewVal string
}

// DiffSecrets compares old and new secret maps and returns a list of changes.
// Only keys present in newSecrets are considered (removed keys are detected from old).
func DiffSecrets(oldSecrets, newSecrets map[string]string) []SecretChange {
	var changes []SecretChange

	// detect added and changed
	for k, newVal := range newSecrets {
		oldVal, exists := oldSecrets[k]
		if !exists {
			changes = append(changes, SecretChange{Key: k, Action: "added", NewVal: newVal})
		} else if oldVal != newVal {
			changes = append(changes, SecretChange{Key: k, Action: "changed", OldVal: oldVal, NewVal: newVal})
		} else {
			changes = append(changes, SecretChange{Key: k, Action: "unchanged", OldVal: oldVal, NewVal: newVal})
		}
	}

	// detect removed
	for k, oldVal := range oldSecrets {
		if _, exists := newSecrets[k]; !exists {
			changes = append(changes, SecretChange{Key: k, Action: "removed", OldVal: oldVal})
		}
	}

	return changes
}

// HasChanges returns true if any change is not "unchanged".
func HasChanges(changes []SecretChange) bool {
	for _, c := range changes {
		if c.Action != "unchanged" {
			return true
		}
	}
	return false
}
