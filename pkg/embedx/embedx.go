package embedx

import (
	"errors"
	"math"
	"sort"
	"sync"
)

type Embedder struct {
	store VectorStore
}

func New(store VectorStore) *Embedder {
	return &Embedder{store: store}
}

type Result struct {
	ID     string
	Score  float32
	Vector []float32
}

func (e *Embedder) Add(id string, vec []float32) error {
	if len(vec) == 0 {
		return errors.New("cannot store empty vector")
	}
	return e.store.SaveVector(id, vec)
}

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
		// dimension mismatch guard
		if len(vec) != len(query) {
			continue
		}

		score := cosineSimilarity(query, vec)

		// skip weird values
		if math.IsNaN(float64(score)) {
			continue
		}

		scores = append(scores, Result{
			ID:     id,
			Score:  score,
			Vector: vec, // later this might become optional for performance
		})
	}

	// no matches? return empty slice, not nil
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

// MemoryStore is an in-memory vector store implementation
type MemoryStore struct {
	data map[string][]float32
	dim  int // dimension requirement, 0 means no restriction
	mu   sync.RWMutex
}

// NewMemoryStore creates a new in-memory vector store
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		data: make(map[string][]float32),
		dim:  0, // no dimension restriction by default
	}
}

// NewMemoryStoreWithDim creates a new in-memory vector store with dimension restriction
func NewMemoryStoreWithDim(dim int) *MemoryStore {
	return &MemoryStore{
		data: make(map[string][]float32),
		dim:  dim,
	}
}

// SaveVector saves a vector with the given ID
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

// GetVector retrieves a vector by ID
func (m *MemoryStore) GetVector(id string) ([]float32, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	vec, exists := m.data[id]
	if !exists {
		return nil, errors.New("vector not found")
	}

	return append([]float32(nil), vec...), nil // copy slice before returning
}

// GetAllVectors returns all stored vectors
func (m *MemoryStore) GetAllVectors() (map[string][]float32, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string][]float32)
	for id, vec := range m.data {
		result[id] = append([]float32(nil), vec...) // copy slice
	}
	return result, nil
}

// Close implements the VectorStore interface (no-op for memory store)
func (m *MemoryStore) Close() error {
	return nil
}
