// Package store exposes the concrete vector store implementations from
// internal/store to public consumers. Go's internal package rule prevents
// external modules from importing internal/store directly, so this package
// provides public wrapper types and constructors for each backend.
//
// The Store interface is the contract every bundled store satisfies. External
// consumers can implement their own stores against this interface and plug
// them into the goembedx ecosystem (embedx.New, search, etc.) without having
// to depend on any internal package.
package store

import "github.com/ldaidone/goembedx/pkg/embedx"

// Store is the interface that all bundled vector stores implement. It covers
// both the basic vector operations (SaveVector, GetVector, GetAllVectors) and
// the full-featured operations with metadata and similarity search (Add, Get,
// Search), plus lifecycle management (Close).
//
// Implementations that satisfy Store automatically satisfy embedx.VectorStore.
type Store interface {
	// SaveVector stores a vector with the given ID without metadata.
	// Returns an error if saving fails.
	SaveVector(id string, vec []float32) error
	// GetVector retrieves a vector by its ID.
	// Returns an error if the vector is not found.
	GetVector(id string) ([]float32, error)
	// GetAllVectors returns all stored vectors.
	// Returns an error if retrieval fails.
	GetAllVectors() (map[string][]float32, error)
	// Add stores a vector with metadata.
	Add(id string, vec []float32, meta map[string]any) error
	// Get retrieves a vector by ID along with its norm and metadata.
	// Returns the vector, its L2 norm, associated metadata, and any error.
	Get(id string) ([]float32, float32, map[string]any, error)
	// Search performs similarity search on stored vectors.
	// Returns the top-k most similar vectors to the query.
	Search(query []float32, k int) ([]embedx.SearchResult, error)
	// Close releases any resources held by the store.
	Close() error
}
