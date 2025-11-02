package badger

import (
	"testing"

	"github.com/ldaidone/goembedx/pkg/embedx"
)

func TestBadgerStoreErrorConditions(t *testing.T) {
	tempDir := t.TempDir()
	store, err := NewBadgerStore(tempDir)
	if err != nil {
		t.Fatalf("NewBadgerStore failed: %v", err)
	}
	defer store.Close()

	// Test operations with invalid/nonexistent data
	_, err = store.GetVector("nonexistent")
	if err == nil {
		t.Error("GetVector should return error for non-existent vector")
	}

	_, _, _, err = store.Get("nonexistent")
	if err == nil {
		t.Error("Get should return error for non-existent vector")
	}
}

func TestBadgerStoreWithRealStoreInterface(t *testing.T) {
	// Verify that BadgerStore properly implements the Store interface
	tempDir := t.TempDir()
	store, err := NewBadgerStore(tempDir)
	if err != nil {
		t.Fatalf("NewBadgerStore failed: %v", err)
	}
	defer store.Close()

	var _ embedx.Store = store
	var _ embedx.VectorStore = store
}

func TestComputeNorm(t *testing.T) {
	tempDir := t.TempDir()
	store, err := NewBadgerStore(tempDir)
	if err != nil {
		t.Fatalf("NewBadgerStore failed: %v", err)
	}
	defer store.Close()

	// Test computeNorm with various vectors
	testCases := []struct {
		vec      []float32
		expected float32
		name     string
	}{
		{[]float32{3, 4}, 5.0, "simple 2D vector (3,4)"},
		{[]float32{1, 0, 0}, 1.0, "unit vector"},
		{[]float32{0, 0, 0}, 0.0, "zero vector"},
		{[]float32{-1, -2, -3}, 3.741657, "negative vector"},
	}

	for _, tc := range testCases {
		norm := store.computeNorm(tc.vec)
		// Allow small floating point differences
		if norm < tc.expected-0.01 || norm > tc.expected+0.01 {
			t.Errorf("%s: expected norm ≈ %f, got %f", tc.name, tc.expected, norm)
		}
	}
}
