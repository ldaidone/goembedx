package store

import (
	"context"

	"github.com/ldaidone/goembedx/internal/store/badger"
	"github.com/ldaidone/goembedx/pkg/embedx"
)

// Badger is a public wrapper around the internal BadgerDB-backed store.
// It implements the Store interface and additionally exposes ImportVectors and
// ExportVectors for bulk operations.
type Badger struct {
	*badger.BadgerStore
}

// NewBadger creates a new BadgerDB-backed store at the given path.
// The path specifies the directory where the database files will be stored.
// Returns an error if the database cannot be opened or initialized.
func NewBadger(path string) (*Badger, error) {
	s, err := badger.NewBadgerStore(path)
	if err != nil {
		return nil, err
	}
	return &Badger{BadgerStore: s}, nil
}

// SearchContext returns the top-k vectors most similar to query, optionally
// restricted by metadata via WithFilter. Results are sorted by descending
// score; the top-k default is 10.
func (s *Badger) SearchContext(ctx context.Context, query []float32, opts ...embedx.SearchOption) ([]embedx.SearchResult, error) {
	cfg := embedx.NewSearchConfig(opts...)
	results, err := s.BadgerStore.Search(query, 0)
	if err != nil {
		return nil, err
	}
	return applySearchConfig(results, cfg), nil
}

// Compile-time interface checks
var _ Store = (*Badger)(nil)
var _ embedx.VectorStore = (*Badger)(nil)
var _ embedx.Store = (*Badger)(nil)
var _ embedx.Searcher = (*Badger)(nil)
