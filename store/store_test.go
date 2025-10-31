package store

import (
	"testing"
)

func TestMemoryStoreAdd(t *testing.T) {
	s := NewMemoryStore(2)
	if err := s.Add("id1", []float32{1, 2}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Len() != 1 {
		t.Fatalf("expected length 1, got %d", s.Len())
	}

	if err := s.Add("bad", []float32{1}); err == nil {
		t.Fatalf("expected dimension mismatch error")
	}
}
