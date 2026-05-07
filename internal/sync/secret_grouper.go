package sync

import (
	"fmt"
	"sort"
	"strings"
)

// GroupBy defines how secrets should be grouped.
type GroupBy string

const (
	GroupByPrefix    GroupBy = "prefix"
	GroupByNamespace GroupBy = "namespace"
)

// SecretGroup holds a named collection of secrets.
type SecretGroup struct {
	Name    string
	Secrets map[string]string
}

// SecretGrouper partitions a flat secret map into named groups.
type SecretGrouper struct {
	by        GroupBy
	delimiter string
}

// NewSecretGrouper creates a SecretGrouper. delimiter is the separator used to
// split keys (e.g. "_" or "/"). Returns an error for unsupported GroupBy values.
func NewSecretGrouper(by GroupBy, delimiter string) (*SecretGrouper, error) {
	if by != GroupByPrefix && by != GroupByNamespace {
		return nil, fmt.Errorf("unsupported group-by value %q: must be %q or %q",
			by, GroupByPrefix, GroupByNamespace)
	}
	if delimiter == "" {
		return nil, fmt.Errorf("delimiter must not be empty")
	}
	return &SecretGrouper{by: by, delimiter: delimiter}, nil
}

// Group partitions secrets by the first segment of each key separated by the
// configured delimiter. Keys with no delimiter are placed in the "default" group.
func (g *SecretGrouper) Group(secrets map[string]string) []SecretGroup {
	grouped := make(map[string]map[string]string)

	for k, v := range secrets {
		parts := strings.SplitN(k, g.delimiter, 2)
		groupName := parts[0]
		if len(parts) == 1 {
			groupName = "default"
		}
		if grouped[groupName] == nil {
			grouped[groupName] = make(map[string]string)
		}
		grouped[groupName][k] = v
	}

	names := make([]string, 0, len(grouped))
	for name := range grouped {
		names = append(names, name)
	}
	sort.Strings(names)

	result := make([]SecretGroup, 0, len(names))
	for _, name := range names {
		result = append(result, SecretGroup{Name: name, Secrets: grouped[name]})
	}
	return result
}
