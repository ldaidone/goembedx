package store

import (
	"context"
	"errors"
	"sort"
	"sync"

	"github.com/ldaidone/goembedx/internal/store/memory"
	"github.com/ldaidone/goembedx/pkg/embedx"
	"github.com/ldaidone/goembedx/vector"
)

// Vector represents a stored vector with its identifier and precomputed norm.
// The norm is stored to enable efficient similarity calculations.
type Vector struct {
	// ID is the unique identifier for this vector.
	ID string
	// Val contains the actual float32 vector data.
	Val []float32
	// Norm is the precomputed L2 norm of the vector.
	Norm float32
}

// Memory is an in-memory vector store backed by the internal memory container.
// It maintains vectors of a fixed dimension, precomputes their norms for fast
// similarity searches, and stores optional metadata alongside each vector.
// Unlike the underlying container, Memory satisfies the Store interface and is
// safe for concurrent use.
type Memory struct {
	// mu guards all access to the underlying container and metadata.
	mu sync.RWMutex
	// store is the internal dimension-constrained vector container.
	store *memory.MemoryStore
	// meta holds the metadata associated with each vector ID.
	meta map[string]map[string]any
}

// NewMemory creates a new in-memory vector store for vectors of the specified
// dimension. The dimension must be greater than 0 and all vectors added to
// this store must match this dimension.
func NewMemory(dim int) *Memory {
	return &Memory{
		store: memory.NewMemoryStore(dim),
		meta:  make(map[string]map[string]any),
	}
}

// Dim returns the dimensionality constraint of this store.
// All vectors in this store have this same dimension.
func (m *Memory) Dim() int { return m.store.Dim() }

// Len returns the number of vectors currently stored in this container.
func (m *Memory) Len() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.store.Len()
}

// Data returns the underlying slice of stored vectors as public Vector values.
// Callers should treat the returned slice as read-only to maintain data
// integrity.
func (m *Memory) Data() []Vector {
	m.mu.RLock()
	defer m.mu.RUnlock()
	raw := m.store.Data()
	vectors := make([]Vector, len(raw))
	for i, v := range raw {
		vectors[i] = Vector{ID: v.ID, Val: v.Val, Norm: v.Norm}
	}
	return vectors
}

// Store interface methods
// SaveVector stores a vector with the given ID without metadata.
// Returns an error if the vector dimension doesn't match the store's dimension
// constraint.
func (m *Memory) SaveVector(id string, vec []float32) error {
	return m.Add(id, vec, nil)
}

// GetVector retrieves a vector by its ID.
// Returns an error if the vector is not found.
func (m *Memory) GetVector(id string) ([]float32, error) {
	vec, _, _, err := m.Get(id)
	if err != nil {
		return nil, err
	}
	return vec, nil
}

// GetAllVectors returns all stored vectors as a map from ID to vector data.
func (m *Memory) GetAllVectors() (map[string][]float32, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	vectors := make(map[string][]float32, m.store.Len())
	for _, v := range m.store.Data() {
		vectors[v.ID] = v.Val
	}
	return vectors, nil
}

// Add stores a vector with the given ID and associated metadata.
// It precomputes the L2 norm of the vector for faster similarity calculations.
// Returns an error if the vector dimension doesn't match the store's dimension
// constraint.
func (m *Memory) Add(id string, vec []float32, meta map[string]any) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := m.store.Add(id, vec); err != nil {
		return err
	}
	m.meta[id] = meta
	return nil
}

// Get retrieves a vector by its ID along with its precomputed norm and
// metadata.
// Returns the vector, its norm, metadata, and any error that occurred.
func (m *Memory) Get(id string) ([]float32, float32, map[string]any, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, v := range m.store.Data() {
		if v.ID == id {
			return v.Val, v.Norm, m.meta[id], nil
		}
	}
	return nil, 0, nil, errors.New("store: vector not found")
}

// Search returns the top-k vectors most similar to the query by cosine
// similarity, computed against the stored precomputed norms. Results are
// sorted by score in descending order.
func (m *Memory) Search(query []float32, k int) ([]embedx.SearchResult, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	queryNorm := vector.Norm(query)
	results := make([]embedx.SearchResult, 0)

	for _, v := range m.store.Data() {
		if len(v.Val) != len(query) || queryNorm == 0 || v.Norm == 0 {
			continue
		}
		score := vector.Dot(query, v.Val) / (queryNorm * v.Norm)
		results = append(results, embedx.SearchResult{
			ID:    v.ID,
			Score: score,
			Meta:  m.meta[v.ID],
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})
	if k > 0 && len(results) > k {
		results = results[:k]
	}
	return results, nil
}

// SearchContext returns the top-k vectors most similar to query, optionally
// restricted by metadata via WithFilter. Results are sorted by descending
// score; the top-k default is 10.
func (m *Memory) SearchContext(ctx context.Context, query []float32, opts ...embedx.SearchOption) ([]embedx.SearchResult, error) {
	cfg := embedx.NewSearchConfig(opts...)
	results, err := m.Search(query, 0)
	if err != nil {
		return nil, err
	}
	return applySearchConfig(results, cfg), nil
}

// Close releases any resources held by the memory store.
// For this in-memory implementation, it's a no-op.
func (m *Memory) Close() error {
	return nil
}

// Compile-time interface checks
var _ Store = (*Memory)(nil)
var _ embedx.VectorStore = (*Memory)(nil)
var _ embedx.Store = (*Memory)(nil)
var _ embedx.Searcher = (*Memory)(nil)
