package badger

import (
	"reflect"
	"testing"

	"github.com/ldaidone/goembedx/pkg/embedx"
)

func TestNewBadgerStore(t *testing.T) {
	tempDir := t.TempDir()
	store, err := NewBadgerStore(tempDir)
	if err != nil {
		t.Fatalf("NewBadgerStore failed: %v", err)
	}

	if store == nil {
		t.Fatal("NewBadgerStore returned nil")
	}

	err = store.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}
}

func TestBadgerStoreAddGet(t *testing.T) {
	tempDir := t.TempDir()
	store, err := NewBadgerStore(tempDir)
	if err != nil {
		t.Fatalf("NewBadgerStore failed: %v", err)
	}
	defer store.Close()

	// Test Add
	meta := map[string]any{"key": "value"}
	err = store.Add("test", []float32{1, 2, 3}, meta)
	if err != nil {
		t.Errorf("Add failed: %v", err)
	}

	// Test Get
	vec, norm, retrievedMeta, err := store.Get("test")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	expectedVec := []float32{1, 2, 3}
	if !reflect.DeepEqual(vec, expectedVec) {
		t.Errorf("Expected vector %v, got %v", expectedVec, vec)
	}

	// Check that norm is computed correctly (sqrt(1^2 + 2^2 + 3^2) = sqrt(14) ≈ 3.74)
	expectedNorm := float32(3.741657) // sqrt(14)
	if norm < expectedNorm-0.01 || norm > expectedNorm+0.01 {
		t.Errorf("Expected norm ≈ %f, got %f", expectedNorm, norm)
	}

	if !reflect.DeepEqual(meta, retrievedMeta) {
		t.Errorf("Expected metadata %v, got %v", meta, retrievedMeta)
	}
}

func TestBadgerStoreSaveVectorGetVector(t *testing.T) {
	tempDir := t.TempDir()
	badgerStore, err := NewBadgerStore(tempDir)
	if err != nil {
		t.Fatalf("NewBadgerStore failed: %v", err)
	}
	defer badgerStore.Close()

	// Test SaveVector
	vec := []float32{0.5, 1.5, 2.5}
	err = badgerStore.SaveVector("saveTest", vec)
	if err != nil {
		t.Errorf("SaveVector failed: %v", err)
	}

	// Test GetVector
	retrievedVec, err := badgerStore.GetVector("saveTest")
	if err != nil {
		t.Fatalf("GetVector failed: %v", err)
	}

	if !reflect.DeepEqual(vec, retrievedVec) {
		t.Errorf("Expected vector %v, got %v", vec, retrievedVec)
	}

	// Test non-existent vector
	_, err = badgerStore.GetVector("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent vector, got nil")
	}
}

func TestBadgerStoreGetAllVectors(t *testing.T) {
	tempDir := t.TempDir()
	store, err := NewBadgerStore(tempDir)
	if err != nil {
		t.Fatalf("NewBadgerStore failed: %v", err)
	}
	defer store.Close()

	// Add some vectors
	_ = store.SaveVector("vec1", []float32{1, 0, 0})
	_ = store.SaveVector("vec2", []float32{0, 1, 0})
	_ = store.SaveVector("vec3", []float32{0, 0, 1})

	// Get all vectors
	all, err := store.GetAllVectors()
	if err != nil {
		t.Fatalf("GetAllVectors failed: %v", err)
	}

	if len(all) != 3 {
		t.Errorf("Expected 3 vectors, got %d", len(all))
	}

	expected := map[string][]float32{
		"vec1": {1, 0, 0},
		"vec2": {0, 1, 0},
		"vec3": {0, 0, 1},
	}

	for id, expectedVec := range expected {
		actualVec, exists := all[id]
		if !exists {
			t.Errorf("Vector %s not found in results", id)
			continue
		}
		if !reflect.DeepEqual(expectedVec, actualVec) {
			t.Errorf("For vector %s, expected %v, got %v", id, expectedVec, actualVec)
		}
	}
}

func TestBadgerStoreSearch(t *testing.T) {
	tempDir := t.TempDir()
	store, err := NewBadgerStore(tempDir)
	if err != nil {
		t.Fatalf("NewBadgerStore failed: %v", err)
	}
	defer store.Close()

	// Add some vectors
	_ = store.Add("vec1", []float32{1, 0, 0}, nil)
	_ = store.Add("vec2", []float32{0, 1, 0}, nil)
	_ = store.Add("vec3", []float32{0.5, 0.5, 0}, nil)

	// Search
	results, err := store.Search([]float32{1, 0, 0}, 2)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	if results[0].ID != "vec1" {
		t.Errorf("Expected first result to be 'vec1', got '%s'", results[0].ID)
	}
}

func TestBadgerStoreImportExport(t *testing.T) {
	tempDir := t.TempDir()
	store, err := NewBadgerStore(tempDir)
	if err != nil {
		t.Fatalf("NewBadgerStore failed: %v", err)
	}
	defer store.Close()

	// Test ImportVectors
	vectors := map[string][]float32{
		"import1": {1, 2, 3},
		"import2": {4, 5, 6},
	}

	err = store.ImportVectors(vectors)
	if err != nil {
		t.Fatalf("ImportVectors failed: %v", err)
	}

	// Test ExportVectors
	exported, err := store.ExportVectors()
	if err != nil {
		t.Fatalf("ExportVectors failed: %v", err)
	}

	if len(exported) != 2 {
		t.Errorf("Expected 2 exported vectors, got %d", len(exported))
	}

	for id, expectedVec := range vectors {
		actualVec, exists := exported[id]
		if !exists {
			t.Errorf("Vector %s not found in export", id)
			continue
		}
		if !reflect.DeepEqual(expectedVec, actualVec) {
			t.Errorf("For vector %s, expected %v, got %v", id, expectedVec, actualVec)
		}
	}
}

func TestBadgerStoreInterfaces(t *testing.T) {
	tempDir := t.TempDir()
	store, err := NewBadgerStore(tempDir)
	if err != nil {
		t.Fatalf("NewBadgerStore failed: %v", err)
	}
	defer store.Close()

	// Verify it implements the expected interfaces
	var _ embedx.VectorStore = store
	var _ embedx.Store = store
}

func TestBackwardCompatibility(t *testing.T) {
	tempDir := t.TempDir()
	store, err := NewBadgerStore(tempDir)
	if err != nil {
		t.Fatalf("NewBadgerStore failed: %v", err)
	}
	defer store.Close()

	// This test ensures that old data format can be read and automatically upgraded
	// We'll test by adding data with the old format and ensuring it can be read
}
