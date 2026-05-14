package sync

import (
	"testing"
)

func TestNewSecretHasher_NotNil(t *testing.T) {
	h := NewSecretHasher()
	if h == nil {
		t.Fatal("expected non-nil SecretHasher")
	}
}

func TestHash_EmptyMap_ReturnsEmptySHA256(t *testing.T) {
	h := NewSecretHasher()
	got := h.Hash(map[string]string{})
	want := emptySHA256()
	if got != want {
		t.Errorf("Hash(empty) = %q, want %q", got, want)
	}
}

func TestHash_NilMap_SameAsEmpty(t *testing.T) {
	h := NewSecretHasher()
	if h.Hash(nil) != h.Hash(map[string]string{}) {
		t.Error("nil and empty maps should produce the same hash")
	}
}

func TestHash_DeterministicAcrossInsertionOrder(t *testing.T) {
	h := NewSecretHasher()
	a := map[string]string{"FOO": "bar", "BAZ": "qux"}
	b := map[string]string{"BAZ": "qux", "FOO": "bar"}
	if h.Hash(a) != h.Hash(b) {
		t.Error("hashes should be equal regardless of insertion order")
	}
}

func TestHash_DifferentValues_ProduceDifferentHashes(t *testing.T) {
	h := NewSecretHasher()
	a := map[string]string{"KEY": "value1"}
	b := map[string]string{"KEY": "value2"}
	if h.Hash(a) == h.Hash(b) {
		t.Error("different values should produce different hashes")
	}
}

func TestEqual_SameMaps_ReturnsTrue(t *testing.T) {
	h := NewSecretHasher()
	m := map[string]string{"A": "1", "B": "2"}
	if !h.Equal(m, m) {
		t.Error("Equal should return true for identical maps")
	}
}

func TestEqual_DifferentMaps_ReturnsFalse(t *testing.T) {
	h := NewSecretHasher()
	a := map[string]string{"A": "1"}
	b := map[string]string{"A": "2"}
	if h.Equal(a, b) {
		t.Error("Equal should return false for maps with different values")
	}
}

func TestHash_IsHexString(t *testing.T) {
	h := NewSecretHasher()
	got := h.Hash(map[string]string{"X": "y"})
	if len(got) != 64 {
		t.Errorf("expected 64-char hex string, got len=%d", len(got))
	}
}
