package sync

import (
	"sort"
	"strings"
)

// SortOrder defines the ordering strategy for secrets.
type SortOrder string

const (
	SortOrderAlpha      SortOrder = "alpha"
	SortOrderAlphaDesc  SortOrder = "alpha_desc"
	SortOrderKeyLength  SortOrder = "key_length"
	SortOrderNone       SortOrder = "none"
)

// ErrInvalidSortOrder is returned when an unrecognised sort order is provided.
type ErrInvalidSortOrder struct {
	Order string
}

func (e *ErrInvalidSortOrder) Error() string {
	return "invalid sort order: " + e.Order
}

// ValidateSortOrder returns an error if the given order is not recognised.
func ValidateSortOrder(order SortOrder) error {
	switch order {
	case SortOrderAlpha, SortOrderAlphaDesc, SortOrderKeyLength, SortOrderNone:
		return nil
	}
	return &ErrInvalidSortOrder{Order: string(order)}
}

// SortSecrets returns a new map whose iteration order is represented as a
// sorted slice of key-value pairs. Because Go maps are unordered the function
// returns the keys in the requested order so callers can iterate deterministically.
func SortSecrets(secrets map[string]string, order SortOrder) ([]string, error) {
	if err := ValidateSortOrder(order); err != nil {
		return nil, err
	}

	keys := make([]string, 0, len(secrets))
	for k := range secrets {
		keys = append(keys, k)
	}

	switch order {
	case SortOrderAlpha:
		sort.Slice(keys, func(i, j int) bool {
			return strings.ToLower(keys[i]) < strings.ToLower(keys[j])
		})
	case SortOrderAlphaDesc:
		sort.Slice(keys, func(i, j int) bool {
			return strings.ToLower(keys[i]) > strings.ToLower(keys[j])
		})
	case SortOrderKeyLength:
		sort.Slice(keys, func(i, j int) bool {
			if len(keys[i]) == len(keys[j]) {
				return keys[i] < keys[j]
			}
			return len(keys[i]) < len(keys[j])
		})
	case SortOrderNone:
		// preserve insertion order is not possible; return as-is
	}

	return keys, nil
}
