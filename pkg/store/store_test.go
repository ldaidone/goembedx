package store

import (
	"reflect"
	"testing"

	"github.com/ldaidone/goembedx/pkg/embedx"
)

func TestNewBadger(t *testing.T) {
	store, err := NewBadger(t.TempDir())
	if err != nil {
		t.Fatalf("NewBadger failed: %v", err)
	}
	if store == nil {
		t.Fatal("NewBadger returned nil")
	}

	var _ Store = store
	var _ embedx.VectorStore = store
	var _ embedx.Store = store

	if err := store.Add("test", []float32{1, 2, 3}, map[string]any{"k": "v"}); err != nil {
		t.Fatalf("Add failed: %v", err)
	}
	vec, norm, meta, err := store.Get("test")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if !reflect.DeepEqual(vec, []float32{1, 2, 3}) {
		t.Errorf("Expected vector %v, got %v", []float32{1, 2, 3}, vec)
	}
	if norm == 0 {
		t.Error("Expected non-zero norm")
	}
	if !reflect.DeepEqual(meta, map[string]any{"k": "v"}) {
		t.Errorf("Expected metadata %v, got %v", map[string]any{"k": "v"}, meta)
	}

	if err := store.Close(); err != nil {
		t.Errorf("Close failed: %v", err)
	}
}

func TestNewSQLite(t *testing.T) {
	store, err := NewSQLite(t.TempDir())
	if err != nil {
		t.Fatalf("NewSQLite failed: %v", err)
	}
	if store == nil {
		t.Fatal("NewSQLite returned nil")
	}

	var _ Store = store
	var _ embedx.VectorStore = store
	var _ embedx.Store = store

	if err := store.Add("test", []float32{1, 0, 0}, nil); err != nil {
		t.Fatalf("Add failed: %v", err)
	}
	results, err := store.Search([]float32{1, 0, 0}, 1)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(results) != 1 || results[0].ID != "test" {
		t.Errorf("Expected 1 result with ID 'test', got %+v", results)
	}

	if err := store.Close(); err != nil {
		t.Errorf("Close failed: %v", err)
	}
}

func TestNewMemory(t *testing.T) {
	m := NewMemory(2)
	if m == nil {
		t.Fatal("NewMemory returned nil")
	}
	if m.Dim() != 2 {
		t.Errorf("Expected dim 2, got %d", m.Dim())
	}

	var _ Store = m
	var _ embedx.VectorStore = m
	var _ embedx.Store = m

	if err := m.Add("id1", []float32{1, 2}, map[string]any{"k": "v"}); err != nil {
		t.Fatalf("Add failed: %v", err)
	}
	if err := m.Add("bad", []float32{1}, nil); err == nil {
		t.Error("Expected dimension mismatch error")
	}
	if m.Len() != 1 {
		t.Errorf("Expected length 1, got %d", m.Len())
	}

	vec, norm, meta, err := m.Get("id1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if !reflect.DeepEqual(vec, []float32{1, 2}) {
		t.Errorf("Expected vector %v, got %v", []float32{1, 2}, vec)
	}
	if norm == 0 {
		t.Error("Expected non-zero norm")
	}
	if !reflect.DeepEqual(meta, map[string]any{"k": "v"}) {
		t.Errorf("Expected metadata %v, got %v", map[string]any{"k": "v"}, meta)
	}

	results, err := m.Search([]float32{1, 2}, 1)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(results) != 1 || results[0].ID != "id1" {
		t.Errorf("Expected 1 result with ID 'id1', got %+v", results)
	}

	all, err := m.GetAllVectors()
	if err != nil {
		t.Fatalf("GetAllVectors failed: %v", err)
	}
	if len(all) != 1 || !reflect.DeepEqual(all["id1"], []float32{1, 2}) {
		t.Errorf("Unexpected GetAllVectors result: %v", all)
	}

	data := m.Data()
	if len(data) != 1 || data[0].ID != "id1" {
		t.Errorf("Unexpected Data() result: %+v", data)
	}
	if !reflect.DeepEqual(data[0].Val, []float32{1, 2}) {
		t.Errorf("Expected Val %v, got %v", []float32{1, 2}, data[0].Val)
	}

	if err := m.Close(); err != nil {
		t.Errorf("Close failed: %v", err)
	}
}
