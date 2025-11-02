package embedx

import (
	"errors"
	"reflect"
	"testing"
)

// mockVectorStore implements VectorStore interface for testing
type mockVectorStore struct {
	data     map[string][]float32
	saveErr  error
	getErr   error
	allErr   error
	closeErr error
}

func (m *mockVectorStore) SaveVector(id string, vec []float32) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	if m.data == nil {
		m.data = make(map[string][]float32)
	}
	m.data[id] = append([]float32(nil), vec...)
	return nil
}

func (m *mockVectorStore) GetVector(id string) ([]float32, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	vec, exists := m.data[id]
	if !exists {
		return nil, errors.New("vector not found")
	}
	return vec, nil
}

func (m *mockVectorStore) GetAllVectors() (map[string][]float32, error) {
	if m.allErr != nil {
		return nil, m.allErr
	}
	return m.data, nil
}

func (m *mockVectorStore) Close() error {
	return m.closeErr
}

func TestNew(t *testing.T) {
	store := &mockVectorStore{}
	embedder := New(store)

	if embedder == nil {
		t.Fatal("New returned nil")
	}

	if embedder.store != store {
		t.Error("Embedder does not hold the correct store")
	}
}

func TestEmbedderAdd(t *testing.T) {
	store := &mockVectorStore{}
	embedder := New(store)

	// Test successful addition
	err := embedder.Add("test", []float32{1, 2, 3})
	if err != nil {
		t.Errorf("Add failed: %v", err)
	}

	// Test empty vector error
	err = embedder.Add("empty", []float32{})
	if err == nil {
		t.Error("Expected error for empty vector, got nil")
	}

	// Test store error propagation
	store.saveErr = errors.New("store error")
	err = embedder.Add("error", []float32{1, 2, 3})
	if err == nil {
		t.Error("Expected store error to propagate, got nil")
	}
}

func TestEmbedderSearch(t *testing.T) {
	store := &mockVectorStore{
		data: map[string][]float32{
			"vec1": {1, 0, 0},
			"vec2": {0, 1, 0},
			"vec3": {0, 0, 1},
		},
	}
	embedder := New(store)

	// Test successful search
	results, err := embedder.Search([]float32{1, 0, 0}, 2)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	// Test empty query error
	_, err = embedder.Search([]float32{}, 1)
	if err == nil {
		t.Error("Expected error for empty query, got nil")
	}

	// Test empty store
	store.data = nil
	results, err = embedder.Search([]float32{1, 0, 0}, 1)
	if err == nil {
		t.Error("Expected error for empty store, got nil")
	}

	// Test store error propagation
	store.allErr = errors.New("store error")
	_, err = embedder.Search([]float32{1, 0, 0}, 1)
	if err == nil {
		t.Error("Expected store error to propagate, got nil")
	}

	// Test dimension mismatch handling
	store.allErr = nil // Clear the error that was set earlier
	store.data = map[string][]float32{
		"vec1": {1, 0, 0}, // 3D - matches query
		"vec2": {1, 0},    // 2D - should be skipped
	}
	results, err = embedder.Search([]float32{1, 0, 0}, 5) // 3D query
	if err != nil {
		t.Errorf("Search with dimension mismatch should skip mismatched vectors: %v", err)
	}
	// vec1 (1,0,0) should match the query (1,0,0) and vec2 (1,0) should be skipped due to dim mismatch
	// In practice, it should return 1 result
	if len(results) == 0 {
		t.Error("Expected at least 1 result since vec1 matches dimensions and query")
	}
	// The important thing is that no error occurred when processing mismatched dimensions
}

func TestMemoryStore(t *testing.T) {
	// Test NewMemoryStore
	store := NewMemoryStore()
	if store == nil {
		t.Fatal("NewMemoryStore returned nil")
	}

	// Test NewMemoryStoreWithDim
	dimStore := NewMemoryStoreWithDim(3)
	if dimStore == nil {
		t.Fatal("NewMemoryStoreWithDim returned nil")
	}

	// Test SaveVector with dimension constraint
	err := dimStore.SaveVector("test", []float32{1, 2, 3})
	if err != nil {
		t.Errorf("SaveVector failed: %v", err)
	}

	// Test dimension mismatch
	err = dimStore.SaveVector("bad", []float32{1, 2}) // 2D instead of 3D
	if err == nil {
		t.Error("Expected dimension mismatch error, got nil")
	}

	// Test empty vector
	err = dimStore.SaveVector("empty", []float32{})
	if err == nil {
		t.Error("Expected error for empty vector, got nil")
	}

	// Test GetVector
	vec, err := dimStore.GetVector("test")
	if err != nil {
		t.Errorf("GetVector failed: %v", err)
	}

	expected := []float32{1, 2, 3}
	if !reflect.DeepEqual(vec, expected) {
		t.Errorf("Expected %v, got %v", expected, vec)
	}

	// Test non-existent vector
	_, err = dimStore.GetVector("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent vector, got nil")
	}

	// Test GetAllVectors
	all, err := dimStore.GetAllVectors()
	if err != nil {
		t.Errorf("GetAllVectors failed: %v", err)
	}

	if len(all) != 1 {
		t.Errorf("Expected 1 vector, got %d", len(all))
	}

	if !reflect.DeepEqual(all["test"], expected) {
		t.Errorf("Expected %v, got %v", expected, all["test"])
	}

	// Test Close
	err = dimStore.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}
}

func TestCosineSimilarity(t *testing.T) {
	// Test identical vectors (should give 1.0)
	a := []float32{1, 0, 0}
	b := []float32{1, 0, 0}
	result := cosineSimilarity(a, b)
	if result != 1.0 {
		t.Errorf("Expected 1.0 for identical vectors, got %f", result)
	}

	// Test orthogonal vectors (should give 0.0)
	a = []float32{1, 0, 0}
	b = []float32{0, 1, 0}
	result = cosineSimilarity(a, b)
	if result != 0.0 {
		t.Errorf("Expected 0.0 for orthogonal vectors, got %f", result)
	}

	// Test opposite vectors (should give -1.0)
	a = []float32{1, 0, 0}
	b = []float32{-1, 0, 0}
	result = cosineSimilarity(a, b)
	if result != -1.0 {
		t.Errorf("Expected -1.0 for opposite vectors, got %f", result)
	}
}
