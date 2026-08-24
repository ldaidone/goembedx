package store

import (
	"context"
	"testing"

	"github.com/ldaidone/goembedx/pkg/embedx"
)

// searcherStore is the subset of the store API exercised by the SearchContext
// tests, implemented by all three bundled backends.
type searcherStore interface {
	embedx.Searcher
	Add(id string, vec []float32, meta map[string]any) error
	Close() error
}

func testSearchContext(t *testing.T, s searcherStore) {
	t.Helper()

	if err := s.Add("doc1", []float32{1, 0, 0}, map[string]any{"type": "doc", "lang": "go"}); err != nil {
		t.Fatalf("Add doc1 failed: %v", err)
	}
	if err := s.Add("doc2", []float32{0, 1, 0}, map[string]any{"type": "entity", "lang": "go"}); err != nil {
		t.Fatalf("Add doc2 failed: %v", err)
	}
	if err := s.Add("doc3", []float32{0.9, 0.1, 0}, map[string]any{"type": "doc", "lang": "rust"}); err != nil {
		t.Fatalf("Add doc3 failed: %v", err)
	}

	ctx := context.Background()
	query := []float32{1, 0, 0}

	// Unfiltered, top-2: doc1 then doc3.
	results, err := s.SearchContext(ctx, query, embedx.WithK(2))
	if err != nil {
		t.Fatalf("SearchContext failed: %v", err)
	}
	if len(results) != 2 || results[0].ID != "doc1" || results[1].ID != "doc3" {
		t.Errorf("Expected [doc1 doc3], got %+v", resultIDs(results))
	}

	// Filtered by type=doc, no k limit: doc1 and doc3, doc2 excluded.
	results, err = s.SearchContext(ctx, query, embedx.WithFilter(embedx.Eq("type", "doc")))
	if err != nil {
		t.Fatalf("SearchContext filtered failed: %v", err)
	}
	if len(results) != 2 || results[0].ID != "doc1" || results[1].ID != "doc3" {
		t.Errorf("Expected filtered [doc1 doc3], got %+v", resultIDs(results))
	}

	// Filter + k: only the top filtered match.
	results, err = s.SearchContext(ctx, query, embedx.WithK(1), embedx.WithFilter(embedx.Eq("lang", "go")))
	if err != nil {
		t.Fatalf("SearchContext filter+k failed: %v", err)
	}
	if len(results) != 1 || results[0].ID != "doc1" {
		t.Errorf("Expected filtered top-1 [doc1], got %+v", resultIDs(results))
	}

	// Composite filter: type=doc AND lang=go → only doc1.
	results, err = s.SearchContext(ctx, query, embedx.WithFilter(embedx.And(
		embedx.Eq("type", "doc"),
		embedx.Eq("lang", "go"),
	)))
	if err != nil {
		t.Fatalf("SearchContext And failed: %v", err)
	}
	if len(results) != 1 || results[0].ID != "doc1" {
		t.Errorf("Expected And result [doc1], got %+v", resultIDs(results))
	}

	// Filter matching nothing returns an empty, non-nil slice.
	results, err = s.SearchContext(ctx, query, embedx.WithFilter(embedx.Eq("type", "missing")))
	if err != nil {
		t.Fatalf("SearchContext no-match failed: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("Expected no results, got %+v", resultIDs(results))
	}

	// Default k is 10: no truncation for 3 vectors.
	results, err = s.SearchContext(ctx, query)
	if err != nil {
		t.Fatalf("SearchContext default failed: %v", err)
	}
	if len(results) != 3 {
		t.Errorf("Expected all 3 results by default, got %d", len(results))
	}
}

func resultIDs(results []embedx.SearchResult) []string {
	ids := make([]string, len(results))
	for i, r := range results {
		ids[i] = r.ID
	}
	return ids
}

func TestBadgerSearchContext(t *testing.T) {
	s, err := NewBadger(t.TempDir())
	if err != nil {
		t.Fatalf("NewBadger failed: %v", err)
	}
	defer s.Close()
	testSearchContext(t, s)
}

func TestSQLiteSearchContext(t *testing.T) {
	s, err := NewSQLite(t.TempDir())
	if err != nil {
		t.Fatalf("NewSQLite failed: %v", err)
	}
	defer s.Close()
	testSearchContext(t, s)
}

func TestMemorySearchContext(t *testing.T) {
	s := NewMemory(3)
	defer s.Close()
	testSearchContext(t, s)
}
