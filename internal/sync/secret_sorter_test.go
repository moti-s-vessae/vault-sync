package sync

import (
	"testing"
)

func TestValidateSortOrder_Valid(t *testing.T) {
	orders := []SortOrder{SortOrderAlpha, SortOrderAlphaDesc, SortOrderKeyLength, SortOrderNone}
	for _, o := range orders {
		if err := ValidateSortOrder(o); err != nil {
			t.Errorf("expected no error for order %q, got %v", o, err)
		}
	}
}

func TestValidateSortOrder_Invalid(t *testing.T) {
	err := ValidateSortOrder(SortOrder("bogus"))
	if err == nil {
		t.Fatal("expected error for invalid sort order, got nil")
	}
	var e *ErrInvalidSortOrder
	switch err.(type) {
	case *ErrInvalidSortOrder:
		e = err.(*ErrInvalidSortOrder)
	default:
		t.Fatalf("unexpected error type: %T", err)
	}
	if e.Order != "bogus" {
		t.Errorf("expected order field %q, got %q", "bogus", e.Order)
	}
}

func TestSortSecrets_Alpha(t *testing.T) {
	secrets := map[string]string{"ZEBRA": "1", "apple": "2", "Mango": "3"}
	keys, err := SortSecrets(secrets, SortOrderAlpha)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"apple", "Mango", "ZEBRA"}
	for i, k := range keys {
		if k != want[i] {
			t.Errorf("index %d: got %q, want %q", i, k, want[i])
		}
	}
}

func TestSortSecrets_AlphaDesc(t *testing.T) {
	secrets := map[string]string{"ZEBRA": "1", "apple": "2", "Mango": "3"}
	keys, err := SortSecrets(secrets, SortOrderAlphaDesc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"ZEBRA", "Mango", "apple"}
	for i, k := range keys {
		if k != want[i] {
			t.Errorf("index %d: got %q, want %q", i, k, want[i])
		}
	}
}

func TestSortSecrets_KeyLength(t *testing.T) {
	secrets := map[string]string{"AB": "1", "A": "2", "ABC": "3"}
	keys, err := SortSecrets(secrets, SortOrderKeyLength)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(keys) != 3 {
		t.Fatalf("expected 3 keys, got %d", len(keys))
	}
	if keys[0] != "A" || keys[1] != "AB" || keys[2] != "ABC" {
		t.Errorf("unexpected order: %v", keys)
	}
}

func TestSortSecrets_None_ReturnsAllKeys(t *testing.T) {
	secrets := map[string]string{"X": "1", "Y": "2", "Z": "3"}
	keys, err := SortSecrets(secrets, SortOrderNone)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(keys) != len(secrets) {
		t.Errorf("expected %d keys, got %d", len(secrets), len(keys))
	}
}

func TestSortSecrets_InvalidOrder_ReturnsError(t *testing.T) {
	_, err := SortSecrets(map[string]string{"K": "v"}, SortOrder("unknown"))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestSortSecrets_EmptyMap(t *testing.T) {
	keys, err := SortSecrets(map[string]string{}, SortOrderAlpha)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(keys) != 0 {
		t.Errorf("expected empty slice, got %v", keys)
	}
}
