package vault

// DiffResult holds the result of comparing two secret maps.
type DiffResult struct {
	Added   map[string]string
	Removed map[string]string
	Changed map[string]string
	Unchanged map[string]string
}

// DiffSecrets compares current secrets (from Vault) against existing secrets
// (e.g. loaded from a .env file) and returns a DiffResult.
func DiffSecrets(current, existing map[string]string) DiffResult {
	result := DiffResult{
		Added:     make(map[string]string),
		Removed:   make(map[string]string),
		Changed:   make(map[string]string),
		Unchanged: make(map[string]string),
	}

	for k, v := range current {
		if old, ok := existing[k]; !ok {
			result.Added[k] = v
		} else if old != v {
			result.Changed[k] = v
		} else {
			result.Unchanged[k] = v
		}
	}

	for k, v := range existing {
		if _, ok := current[k]; !ok {
			result.Removed[k] = v
		}
	}

	return result
}

// HasChanges returns true if there are any added, removed, or changed secrets.
func (d DiffResult) HasChanges() bool {
	return len(d.Added) > 0 || len(d.Removed) > 0 || len(d.Changed) > 0
}
