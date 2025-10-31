package search

import (
	"testing"

	"github.com/ldaidone/goembedx/store"
)

func TestBruteSearchBasic(t *testing.T) {
	s := store.NewMemoryStore(3)
	_ = s.Add("a", []float32{1, 0, 0}) // similar to query
	_ = s.Add("b", []float32{0, 1, 0})
	_ = s.Add("c", []float32{0, 0, 1})

	query := []float32{1, 0, 0}
	res := SearchBrute(s, query, 2)
	if len(res) != 2 {
		t.Fatalf("expected 2 results, got %d", len(res))
	}
	if res[0].ID != "a" {
		t.Fatalf("expected top result 'a', got %s", res[0].ID)
	}
}

func TestBruteSearchAll(t *testing.T) {
	s := store.NewMemoryStore(2)
	_ = s.Add("x", []float32{1, 0})
	_ = s.Add("y", []float32{0, 1})

	res := SearchBrute(s, []float32{1, 0}, 0)
	if len(res) != 2 {
		t.Fatalf("expected 2 results when k=0, got %d", len(res))
	}
	if res[0].ID != "x" {
		t.Fatalf("expected x first, got %s", res[0].ID)
	}
}
