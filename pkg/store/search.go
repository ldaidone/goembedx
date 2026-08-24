package store

import (
	"github.com/ldaidone/goembedx/pkg/embedx"
)

// applySearchConfig narrows a fully-ranked result list (as produced by the
// bundled stores' brute-force Search) by the resolved search configuration:
// metadata filtering first, then top-k truncation. Results are already sorted
// by descending score.
func applySearchConfig(results []embedx.SearchResult, cfg embedx.SearchConfig) []embedx.SearchResult {
	if cfg.Filter != nil {
		filtered := results[:0]
		for _, r := range results {
			if cfg.Filter(r.Meta) {
				filtered = append(filtered, r)
			}
		}
		results = filtered
	}
	if cfg.K > 0 && len(results) > cfg.K {
		results = results[:cfg.K]
	}
	return results
}
