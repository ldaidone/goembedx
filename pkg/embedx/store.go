package embedx

type SearchResult struct {
	ID    string
	Score float32
	Meta  map[string]any
}

type VectorStore interface {
	SaveVector(id string, vec []float32) error
	GetVector(id string) ([]float32, error)
	GetAllVectors() (map[string][]float32, error)
	Close() error
}

// Store is the full-featured store interface including metadata and search
type Store interface {
	Add(id string, vec []float32, meta map[string]any) error
	Get(id string) ([]float32, float32, map[string]any, error)
	Search(query []float32, k int) ([]SearchResult, error)
	Close() error
}
