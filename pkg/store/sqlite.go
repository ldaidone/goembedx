package store

import (
	"context"

	"github.com/ldaidone/goembedx/internal/store/sqlite"
	"github.com/ldaidone/goembedx/pkg/embedx"
)

// SQLite is a public wrapper around the internal SQLite-backed store.
// It implements the Store interface and additionally exposes ImportVectors,
// ExportVectors, and DeleteStale.
type SQLite struct {
	*sqlite.SQLiteStore
}

// NewSQLite creates a new SQLite-backed store in the given directory.
// The database file (vectors.db) is created inside dir, which is created when
// it does not exist.
// Returns an error if the database cannot be opened or initialized.
func NewSQLite(dir string) (*SQLite, error) {
	s, err := sqlite.NewSQLiteStore(dir)
	if err != nil {
		return nil, err
	}
	return &SQLite{SQLiteStore: s}, nil
}

// SearchContext returns the top-k vectors most similar to query, optionally
// restricted by metadata via WithFilter. Results are sorted by descending
// score; the top-k default is 10.
func (s *SQLite) SearchContext(ctx context.Context, query []float32, opts ...embedx.SearchOption) ([]embedx.SearchResult, error) {
	cfg := embedx.NewSearchConfig(opts...)
	results, err := s.SQLiteStore.Search(query, 0)
	if err != nil {
		return nil, err
	}
	return applySearchConfig(results, cfg), nil
}

// Compile-time interface checks
var _ Store = (*SQLite)(nil)
var _ embedx.VectorStore = (*SQLite)(nil)
var _ embedx.Store = (*SQLite)(nil)
var _ embedx.Searcher = (*SQLite)(nil)
