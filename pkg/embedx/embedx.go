// Package embedx provides core vector embedding storage functionality.
// It implements the core logic for vector storage, retrieval, and similarity search.
package embedx

import (
	"errors"
	"math"
	"sort"
	"sync"
)

// Embedder provides the core functionality for adding and searching vectors.
type Embedder struct {
	// store holds the underlying vector storage implementation.
	store VectorStore
}

// New creates a new Embedder instance with the specified vector store.
// The store must implement the VectorStore interface and handle the actual
// storage and retrieval of vectors.
func New(store VectorStore) *Embedder {
	return &Embedder{store: store}
}

// Result represents a single search result with ID, similarity score, and vector data.
type Result struct {
	// ID is the identifier of the matching vector.
	ID string
	// Score is the cosine similarity score between -1.0 and 1.0.
	// Higher scores indicate greater similarity.
	Score float32
	// Vector contains the actual vector data of the result.
	Vector []float32
}

// Add adds a vector with the specified ID to the store.
// It returns an error if the vector is empty or if the underlying store returns an error.
func (e *Embedder) Add(id string, vec []float32) error {
	if len(vec) == 0 {
		return errors.New("cannot store empty vector")
	}
	return e.store.SaveVector(id, vec)
}

// Search performs a similarity search against all stored vectors.
// It computes cosine similarity between the query vector and all stored vectors,
// then returns the top-k most similar results sorted by score in descending order.
//
// Returns an error if the query vector is empty, the store is empty, or if
// the underlying store returns an error during retrieval.
func (e *Embedder) Search(query []float32, k int) ([]Result, error) {
	if len(query) == 0 {
		return nil, errors.New("query vector is empty")
	}

	items, err := e.store.GetAllVectors()
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, errors.New("vector store is empty")
	}

	scores := make([]Result, 0, len(items))

	for id, vec := range items {
		// Skip vectors with mismatched dimensions
		if len(vec) != len(query) {
			continue
		}

		score := cosineSimilarity(query, vec)

		// Skip results with NaN scores
		if math.IsNaN(float64(score)) {
			continue
		}

		scores = append(scores, Result{
			ID:     id,
			Score:  score,
			Vector: vec, // later this might become optional for performance
		})
	}

	// Return empty slice if no matches found
	if len(scores) == 0 {
		return []Result{}, nil
	}

	sort.Slice(scores, func(i, j int) bool {
		return scores[i].Score > scores[j].Score
	})

	if len(scores) > k {
		scores = scores[:k]
	}

	return scores, nil
}

// cosineSimilarity computes the cosine similarity between two vectors.
// It returns a value between -1.0 and 1.0 indicating the cosine of the angle between vectors.
// Values closer to 1.0 indicate higher similarity.
func cosineSimilarity(a, b []float32) float32 {
	var dot, magA, magB float32

	for i := range a {
		dot += a[i] * b[i]
		magA += a[i] * a[i]
		magB += b[i] * b[i]
	}

	den := float32(math.Sqrt(float64(magA)) * math.Sqrt(float64(magB)))
	if den == 0 {
		return 0
	}

	return dot / den
}

// MemoryStore implements an in-memory vector store with thread-safe operations.
// It optionally enforces dimension constraints on stored vectors.
type MemoryStore struct {
	// data holds the map of vector IDs to vector data.
	data map[string][]float32
	// dim specifies the required dimension for stored vectors.
	// If 0, no dimension restriction is enforced.
	dim int
	// mu provides read-write mutex for thread-safe access to data.
	mu sync.RWMutex
}

// NewMemoryStore creates a new in-memory vector store with no dimension restriction.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		data: make(map[string][]float32),
		dim:  0, // no dimension restriction by default
	}
}

// NewMemoryStoreWithDim creates a new in-memory vector store with the specified dimension restriction.
// All vectors stored in this store must have the given dimension.
func NewMemoryStoreWithDim(dim int) *MemoryStore {
	return &MemoryStore{
		data: make(map[string][]float32),
		dim:  dim,
	}
}

// SaveVector stores a vector with the given ID.
// Returns an error if the ID is empty, the vector is empty,
// or if the vector dimension doesn't match the store's dimension requirement.
func (m *MemoryStore) SaveVector(id string, vec []float32) error {
	if id == "" {
		return errors.New("id cannot be empty")
	}
	if len(vec) == 0 {
		return errors.New("vector cannot be empty")
	}

	if m.dim > 0 && len(vec) != m.dim {
		return errors.New("vector dimension mismatch")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.data[id] = append([]float32(nil), vec...) // copy slice to avoid external mutation
	return nil
}

// GetVector retrieves a vector by its ID.
// Returns an error if the vector is not found in the store.
func (m *MemoryStore) GetVector(id string) ([]float32, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	vec, exists := m.data[id]
	if !exists {
		return nil, errors.New("vector not found")
	}

	return append([]float32(nil), vec...), nil // copy slice before returning
}

// GetAllVectors returns all stored vectors as a map from ID to vector data.
func (m *MemoryStore) GetAllVectors() (map[string][]float32, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string][]float32)
	for id, vec := range m.data {
		result[id] = append([]float32(nil), vec...) // copy slice
	}
	return result, nil
}

// Close releases any resources held by the memory store.
// For this in-memory implementation, it's a no-op.
func (m *MemoryStore) Close() error {
	return nil
}
