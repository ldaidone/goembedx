// Package embedx provides core vector embedding storage functionality.
package embedx

// SearchResult represents a single search result with ID, score, and metadata.
type SearchResult struct {
	// ID is the identifier of the matching vector.
	ID string
	// Score is the similarity score for this result.
	Score float32
	// Meta contains additional metadata associated with the vector.
	Meta map[string]any
}

// VectorStore defines the interface for basic vector storage operations.
// It provides methods for storing, retrieving, and managing vectors.
type VectorStore interface {
	// SaveVector stores a vector with the given ID.
	// Returns an error if saving fails.
	SaveVector(id string, vec []float32) error
	// GetVector retrieves a vector by its ID.
	// Returns an error if the vector is not found.
	GetVector(id string) ([]float32, error)
	// GetAllVectors returns all stored vectors.
	// Returns an error if retrieval fails.
	GetAllVectors() (map[string][]float32, error)
	// Close releases any resources held by the store.
	Close() error
}

// Store defines the full-featured store interface with metadata and search capabilities.
// It extends basic vector storage with metadata support and search functionality.
type Store interface {
	// Add stores a vector with metadata.
	Add(id string, vec []float32, meta map[string]any) error
	// Get retrieves a vector by ID along with its norm and metadata.
	// Returns the vector, its L2 norm, associated metadata, and any error.
	Get(id string) ([]float32, float32, map[string]any, error)
	// Search performs similarity search on stored vectors.
	// Returns the top-k most similar vectors to the query.
	Search(query []float32, k int) ([]SearchResult, error)
	// Close releases any resources held by the store.
	Close() error
}
