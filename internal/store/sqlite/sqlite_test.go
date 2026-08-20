package sqlite

import (
	"bytes"
	"encoding/gob"
	"reflect"
	"testing"

	"github.com/ldaidone/goembedx/pkg/embedx"
)

func TestNewSQLiteStore(t *testing.T) {
	tempDir := t.TempDir()
	store, err := NewSQLiteStore(tempDir)
	if err != nil {
		t.Fatalf("NewSQLiteStore failed: %v", err)
	}

	if store == nil {
		t.Fatal("NewSQLiteStore returned nil")
	}

	err = store.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}
}

func TestSQLiteStoreAddGet(t *testing.T) {
	tempDir := t.TempDir()
	store, err := NewSQLiteStore(tempDir)
	if err != nil {
		t.Fatalf("NewSQLiteStore failed: %v", err)
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

func TestSQLiteStoreSaveVectorGetVector(t *testing.T) {
	tempDir := t.TempDir()
	store, err := NewSQLiteStore(tempDir)
	if err != nil {
		t.Fatalf("NewSQLiteStore failed: %v", err)
	}
	defer store.Close()

	// Test SaveVector
	vec := []float32{0.5, 1.5, 2.5}
	err = store.SaveVector("saveTest", vec)
	if err != nil {
		t.Errorf("SaveVector failed: %v", err)
	}

	// Test GetVector
	retrievedVec, err := store.GetVector("saveTest")
	if err != nil {
		t.Fatalf("GetVector failed: %v", err)
	}

	if !reflect.DeepEqual(vec, retrievedVec) {
		t.Errorf("Expected vector %v, got %v", vec, retrievedVec)
	}

	// Test non-existent vector
	_, err = store.GetVector("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent vector, got nil")
	}
}

func TestSQLiteStoreUpsert(t *testing.T) {
	tempDir := t.TempDir()
	store, err := NewSQLiteStore(tempDir)
	if err != nil {
		t.Fatalf("NewSQLiteStore failed: %v", err)
	}
	defer store.Close()

	if err := store.SaveVector("dup", []float32{1, 2, 3}); err != nil {
		t.Fatalf("SaveVector failed: %v", err)
	}
	if err := store.SaveVector("dup", []float32{4, 5, 6}); err != nil {
		t.Fatalf("SaveVector (overwrite) failed: %v", err)
	}

	vec, err := store.GetVector("dup")
	if err != nil {
		t.Fatalf("GetVector failed: %v", err)
	}
	if !reflect.DeepEqual(vec, []float32{4, 5, 6}) {
		t.Errorf("Expected overwritten vector %v, got %v", []float32{4, 5, 6}, vec)
	}
}

func TestSQLiteStoreGetAllVectors(t *testing.T) {
	tempDir := t.TempDir()
	store, err := NewSQLiteStore(tempDir)
	if err != nil {
		t.Fatalf("NewSQLiteStore failed: %v", err)
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

func TestSQLiteStoreSearch(t *testing.T) {
	tempDir := t.TempDir()
	store, err := NewSQLiteStore(tempDir)
	if err != nil {
		t.Fatalf("NewSQLiteStore failed: %v", err)
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

func TestSQLiteStoreImportExport(t *testing.T) {
	tempDir := t.TempDir()
	store, err := NewSQLiteStore(tempDir)
	if err != nil {
		t.Fatalf("NewSQLiteStore failed: %v", err)
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

func TestSQLiteStoreDeleteStale(t *testing.T) {
	tempDir := t.TempDir()
	store, err := NewSQLiteStore(tempDir)
	if err != nil {
		t.Fatalf("NewSQLiteStore failed: %v", err)
	}
	defer store.Close()

	_ = store.SaveVector("keep1", []float32{1, 0, 0})
	_ = store.SaveVector("keep2", []float32{0, 1, 0})
	_ = store.SaveVector("stale1", []float32{0, 0, 1})

	deleted, err := store.DeleteStale(map[string]struct{}{"keep1": {}, "keep2": {}})
	if err != nil {
		t.Fatalf("DeleteStale failed: %v", err)
	}
	if deleted != 1 {
		t.Errorf("Expected 1 deleted vector, got %d", deleted)
	}

	all, err := store.GetAllVectors()
	if err != nil {
		t.Fatalf("GetAllVectors failed: %v", err)
	}
	if len(all) != 2 {
		t.Errorf("Expected 2 remaining vectors, got %d", len(all))
	}
	if _, exists := all["stale1"]; exists {
		t.Error("stale1 should have been deleted")
	}
}

func TestSQLiteStoreInterfaces(t *testing.T) {
	tempDir := t.TempDir()
	store, err := NewSQLiteStore(tempDir)
	if err != nil {
		t.Fatalf("NewSQLiteStore failed: %v", err)
	}
	defer store.Close()

	// Verify it implements the expected interfaces
	var _ embedx.VectorStore = store
	var _ embedx.Store = store
}

func TestSQLiteStoreErrorConditions(t *testing.T) {
	tempDir := t.TempDir()
	store, err := NewSQLiteStore(tempDir)
	if err != nil {
		t.Fatalf("NewSQLiteStore failed: %v", err)
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

func TestSQLiteStoreBackwardCompatibility(t *testing.T) {
	tempDir := t.TempDir()
	store, err := NewSQLiteStore(tempDir)
	if err != nil {
		t.Fatalf("NewSQLiteStore failed: %v", err)
	}
	defer store.Close()

	// Simulate a record written by the legacy []float32-only format
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode([]float32{1, 2, 3}); err != nil {
		t.Fatalf("failed to encode legacy vector: %v", err)
	}
	if _, err := store.db.Exec(`INSERT INTO vectors(id, data) VALUES(?, ?)`, "legacy", buf.Bytes()); err != nil {
		t.Fatalf("failed to insert legacy record: %v", err)
	}

	// GetVector should transparently migrate the legacy format
	vec, err := store.GetVector("legacy")
	if err != nil {
		t.Fatalf("GetVector failed on legacy record: %v", err)
	}
	if !reflect.DeepEqual(vec, []float32{1, 2, 3}) {
		t.Errorf("Expected legacy vector %v, got %v", []float32{1, 2, 3}, vec)
	}

	// Get should expose the computed norm and nil metadata
	_, norm, meta, err := store.Get("legacy")
	if err != nil {
		t.Fatalf("Get failed on legacy record: %v", err)
	}
	expectedNorm := float32(3.741657) // sqrt(14)
	if norm < expectedNorm-0.01 || norm > expectedNorm+0.01 {
		t.Errorf("Expected norm ≈ %f, got %f", expectedNorm, norm)
	}
	if meta != nil {
		t.Errorf("Expected nil metadata for legacy record, got %v", meta)
	}
}
