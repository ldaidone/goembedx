package store

import (
	"errors"
	"github.com/ldaidone/goembedx/vector"
)

// Vector is a stored vector with an ID.
type Vector struct {
	ID   string
	Val  []float32
	Norm float32
}

// MemoryStore is a very small in-memory vector container optimized for read-heavy workloads.
// It is not thread-safe; callers should synchronize if used concurrently.
type MemoryStore struct {
	dim  int
	data []Vector
}

// NewMemoryStore creates a store for vectors of dimension dim. dim must be > 0.
func NewMemoryStore(dim int) *MemoryStore {
	return &MemoryStore{
		dim:  dim,
		data: make([]Vector, 0),
	}
}

// Dim returns the dimensionality of this store.
func (s *MemoryStore) Dim() int { return s.dim }

// Add inserts a vector into the store. It returns an error if the vector length doesn't match store dim.
func (s *MemoryStore) Add(id string, vec []float32) error {
	if len(vec) != s.dim {
		return errors.New("store: vector dimension mismatch")
	}
	n := vector.Norm(vec)
	s.data = append(s.data, Vector{ID: id, Val: vec, Norm: n})
	return nil
}

// Data returns the underlying slice of vectors (read-only semantics expected).
func (s *MemoryStore) Data() []Vector {
	return s.data
}

// Len returns number of vectors stored.
func (s *MemoryStore) Len() int { return len(s.data) }
