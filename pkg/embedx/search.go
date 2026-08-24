// Search options, filters, and the Searcher contract.
//
// Searcher is the search abstraction that both brute-force and future ANN
// (e.g. HNSW) implementations satisfy, so callers and higher-level helpers can
// stay agnostic to the underlying retrieval strategy.
package embedx

import (
	"context"
	"reflect"
)

// Searcher defines the search contract over a vector store. Implementations
// return the top-k results for a query, optionally restricted to vectors whose
// metadata satisfies a Filter (see WithFilter).
//
// The method is named SearchContext (following the database/sql convention) so
// that implementations can also expose the legacy Store.Search(query, k)
// signature without a name collision.
type Searcher interface {
	// SearchContext returns the most similar vectors to query, resolved from
	// the supplied options (e.g. WithK, WithFilter).
	SearchContext(ctx context.Context, query []float32, opts ...SearchOption) ([]SearchResult, error)
}

// SearchOption configures a search. Options are applied in order and mutate a
// SearchConfig resolved by NewSearchConfig.
type SearchOption func(*SearchConfig)

// SearchConfig holds the resolved parameters of a search.
type SearchConfig struct {
	// K is the maximum number of results to return. 0 means all matching
	// results.
	K int
	// Filter, when non-nil, restricts results to vectors whose metadata
	// satisfies the predicate. Applied before ranking/truncation.
	Filter Filter
}

// NewSearchConfig resolves an option list into a SearchConfig, applying
// defaults for any option that was not provided.
func NewSearchConfig(opts ...SearchOption) SearchConfig {
	cfg := SearchConfig{K: 10}
	for _, opt := range opts {
		opt(&cfg)
	}
	return cfg
}

// WithK returns a SearchOption that limits results to the top-k matches.
// k <= 0 means no limit.
func WithK(k int) SearchOption {
	return func(cfg *SearchConfig) { cfg.K = k }
}

// WithFilter returns a SearchOption that only returns vectors whose metadata
// satisfies the given filter.
func WithFilter(f Filter) SearchOption {
	return func(cfg *SearchConfig) { cfg.Filter = f }
}

// Filter decides whether a vector's metadata qualifies for search results.
type Filter func(meta map[string]any) bool

// And returns a Filter that matches when all of the given filters match.
func And(filters ...Filter) Filter {
	return func(meta map[string]any) bool {
		for _, f := range filters {
			if !f(meta) {
				return false
			}
		}
		return true
	}
}

// Or returns a Filter that matches when any of the given filters match.
func Or(filters ...Filter) Filter {
	return func(meta map[string]any) bool {
		for _, f := range filters {
			if f(meta) {
				return true
			}
		}
		return false
	}
}

// Not returns a Filter that matches when the given filter does not match.
func Not(f Filter) Filter {
	return func(meta map[string]any) bool { return !f(meta) }
}

// Eq returns a Filter matching vectors whose metadata has key set to value.
func Eq(key string, value any) Filter {
	return func(meta map[string]any) bool {
		if meta == nil {
			return false
		}
		v, ok := meta[key]
		return ok && reflect.DeepEqual(v, value)
	}
}

// In returns a Filter matching vectors whose metadata value for key is one of
// the given values.
func In(key string, values ...any) Filter {
	return func(meta map[string]any) bool {
		if meta == nil || len(values) == 0 {
			return false
		}
		v, ok := meta[key]
		if !ok {
			return false
		}
		for _, candidate := range values {
			if reflect.DeepEqual(v, candidate) {
				return true
			}
		}
		return false
	}
}

// Exists returns a Filter matching vectors whose metadata contains the given
// key.
func Exists(key string) Filter {
	return func(meta map[string]any) bool {
		if meta == nil {
			return false
		}
		_, ok := meta[key]
		return ok
	}
}
