package ann

import (
	"errors"
	"math/rand"
	"testing"
)

func TestEmptyIndex(t *testing.T) {
	idx := New()
	if idx.Len() != 0 {
		t.Errorf("Expected Len 0, got %d", idx.Len())
	}
	got := idx.Search([]float32{1, 0, 0}, 5)
	if got == nil || len(got) != 0 {
		t.Errorf("Expected empty search result, got %+v", got)
	}
}

func TestAddAndSearch(t *testing.T) {
	idx := New(WithM(8), WithEfConstruction(50), WithEfSearch(50))
	vectors := [][]float32{
		{1, 0, 0},
		{0, 1, 0},
		{0, 0, 1},
		{0.9, 0.1, 0},
	}
	for i, v := range vectors {
		if err := idx.Add(string(rune('a'+i)), v); err != nil {
			t.Fatalf("Add %d failed: %v", i, err)
		}
	}
	if idx.Len() != 4 {
		t.Errorf("Expected Len 4, got %d", idx.Len())
	}

	got := idx.Search([]float32{1, 0, 0}, 3)
	if len(got) != 3 {
		t.Fatalf("Expected 3 results, got %d: %+v", len(got), got)
	}
	if got[0].ID != "a" {
		t.Errorf("Expected nearest neighbor 'a', got %q", got[0].ID)
	}
	// Scores must be descending cosine similarity.
	for i := 1; i < len(got); i++ {
		if got[i].Score > got[i-1].Score {
			t.Errorf("Scores not descending: %+v", got)
		}
	}
}

func TestAddEmptyVector(t *testing.T) {
	idx := New()
	if err := idx.Add("x", nil); err == nil {
		t.Error("Expected error for empty vector")
	}
	if err := idx.Add("x", []float32{}); err == nil {
		t.Error("Expected error for empty vector")
	}
}

func TestDimensionMismatch(t *testing.T) {
	idx := New()
	if err := idx.Add("a", []float32{1, 2, 3}); err != nil {
		t.Fatalf("first Add failed: %v", err)
	}
	err := idx.Add("b", []float32{1, 2})
	if err == nil {
		t.Fatal("Expected dimension mismatch error")
	}
	if !errors.Is(err, ErrDimensionMismatch) {
		t.Errorf("Expected ErrDimensionMismatch, got %v", err)
	}
}

func TestDeterministicAcrossInstances(t *testing.T) {
	build := func() []Result {
		idx := New()
		rng := rand.New(rand.NewSource(7))
		for i := 0; i < 200; i++ {
			v := []float32{rng.Float32()*2 - 1, rng.Float32()*2 - 1, rng.Float32()*2 - 1}
			if err := idx.Add(string(rune('a'+i%26)), v); err != nil {
				t.Fatalf("Add failed: %v", err)
			}
		}
		return idx.Search([]float32{1, 0.5, 0.25}, 10)
	}
	a := build()
	b := build()
	if len(a) != len(b) {
		t.Fatalf("Result count differs: %d vs %d", len(a), len(b))
	}
	for i := range a {
		if a[i].ID != b[i].ID || a[i].Score != b[i].Score {
			t.Errorf("Result %d differs: %+v vs %+v", i, a[i], b[i])
		}
	}
}

func TestSearchRecall(t *testing.T) {
	// With a well-connected index and a forgiving beam width, the exact
	// nearest neighbor should be recovered on a small clustered dataset.
	idx := New(WithEfSearch(100))
	rng := rand.New(rand.NewSource(3))
	// Cluster near (1,0,0).
	for i := 0; i < 50; i++ {
		v := []float32{1 + 0.1*rng.Float32(), 0.1 * rng.Float32(), 0.1 * rng.Float32()}
		if err := idx.Add("near", v); err != nil {
			t.Fatalf("Add failed: %v", err)
		}
	}
	// Distant points.
	for i := 0; i < 50; i++ {
		v := []float32{0.1 * rng.Float32(), 0.1 * rng.Float32(), 1 + 0.1*rng.Float32()}
		if err := idx.Add("far", v); err != nil {
			t.Fatalf("Add failed: %v", err)
		}
	}
	got := idx.Search([]float32{1, 0, 0}, 10)
	for i, r := range got {
		if r.ID != "near" {
			t.Errorf("Result %d should be 'near', got %q", i, r.ID)
		}
	}
}

func TestSearchKClamping(t *testing.T) {
	idx := New()
	if err := idx.Add("a", []float32{1, 0, 0}); err != nil {
		t.Fatalf("Add failed: %v", err)
	}
	if err := idx.Add("b", []float32{0, 1, 0}); err != nil {
		t.Fatalf("Add failed: %v", err)
	}
	if got := idx.Search([]float32{1, 0, 0}, 100); len(got) != 2 {
		t.Errorf("Expected 2 results when k exceeds size, got %d", len(got))
	}
	if got := idx.Search([]float32{1, 0, 0}, 0); len(got) != 0 {
		t.Errorf("Expected 0 results for k=0, got %d", len(got))
	}
}

func TestInterfaceConformance(t *testing.T) {
	var _ Index = (*HNSW)(nil)
}
