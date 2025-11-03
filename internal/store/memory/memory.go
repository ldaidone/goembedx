// Package memory provides an in-memory vector store implementation.
// This implementation stores vectors in memory for fast access, suitable for smaller datasets
// or testing environments where persistence is not required.
package memory

import (
	"errors"
	"github.com/ldaidone/goembedx/vector"
)

// Vector represents a stored vector with its identifier and precomputed norm.
// The norm is stored to enable efficient similarity calculations.
type Vector struct {
	// ID is the unique identifier for this vector.
	ID string
	// Val contains the actual float32 vector data.
	Val []float32
	// Norm is the precomputed L2 norm of the vector for efficient similarity calculations.
	Norm float32
}

// MemoryStore is an in-memory vector container optimized for read-heavy workloads.
// It maintains vectors of fixed dimension and precomputes their norms for fast similarity searches.
// Note: This implementation is not thread-safe; callers must synchronize access if used concurrently.
type MemoryStore struct {
	// dim specifies the required dimension for all vectors in this store.
	dim int
	// data contains the slice of stored vectors.
	data []Vector
}

// NewMemoryStore creates a new in-memory vector store for vectors of the specified dimension.
// The dimension must be greater than 0 and all vectors added to this store must match this dimension.
func NewMemoryStore(dim int) *MemoryStore {
	return &MemoryStore{
		dim:  dim,
		data: make([]Vector, 0),
	}
}

// Dim returns the dimensionality constraint of this store.
// All vectors in this store have this same dimension.
func (s *MemoryStore) Dim() int { return s.dim }

// Add inserts a vector with the given ID into the store.
// It precomputes the L2 norm of the vector for efficient similarity calculations.
// Returns an error if the vector dimension doesn't match the store's dimension constraint.
func (s *MemoryStore) Add(id string, vec []float32) error {
	if len(vec) != s.dim {
		return errors.New("store: vector dimension mismatch")
	}
	n := vector.Norm(vec)
	s.data = append(s.data, Vector{ID: id, Val: vec, Norm: n})
	return nil
}

// Data returns the underlying slice of stored vectors.
// Callers should treat the returned slice as read-only to maintain data integrity.
func (s *MemoryStore) Data() []Vector {
	return s.data
}

// Len returns the number of vectors currently stored in this container.
func (s *MemoryStore) Len() int { return len(s.data) }
