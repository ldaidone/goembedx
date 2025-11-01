package badger

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/dgraph-io/badger/v4"
	"github.com/ldaidone/goembedx/pkg/embedx" // only for the interface
	"math"
	"sort"
)

type BadgerStore struct {
	db *badger.DB
}

// Compile-time interface checks
var _ embedx.VectorStore = (*BadgerStore)(nil)
var _ embedx.Store = (*BadgerStore)(nil)

func NewBadgerStore(path string) (*BadgerStore, error) {
	opts := badger.DefaultOptions(path).WithLogger(nil)
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}
	return &BadgerStore{db: db}, nil
}

// VectorStore interface methods
func (s *BadgerStore) SaveVector(id string, vec []float32) error {
	// Use the same data structure as Add to maintain consistency
	var norm float32
	for _, val := range vec {
		norm += val * val
	}
	norm = float32(math.Sqrt(float64(norm)))

	data := vectorData{
		Vector: vec,
		Norm:   norm,
		Meta:   nil, // No metadata for basic SaveVector
	}

	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(data); err != nil {
		return err
	}

	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(id), buf.Bytes())
	})
}

func (s *BadgerStore) GetVector(id string) ([]float32, error) {
	var data vectorData

	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(id))
		if err != nil {
			return err
		}
		return item.Value(func(v []byte) error {
			// First try to decode as the new vectorData struct
			dec := gob.NewDecoder(bytes.NewReader(v))
			err := dec.Decode(&data)
			if err != nil {
				// If that fails, try to decode as the old []float32 format
				var oldVec []float32
				decOld := gob.NewDecoder(bytes.NewReader(v))
				oldErr := decOld.Decode(&oldVec)
				if oldErr != nil {
					return fmt.Errorf("failed to decode vector data: %w", err)
				}
				// Convert to new format with computed norm
				var norm float32
				for _, val := range oldVec {
					norm += val * val
				}
				norm = float32(math.Sqrt(float64(norm)))
				data = vectorData{
					Vector: oldVec,
					Norm:   norm,
					Meta:   nil,
				}
				// Optionally update the stored record to new format
				_ = s.updateToNewFormat(id, &data) // Don't fail if this fails
			}
			return nil
		})
	})

	if err != nil {
		return nil, err
	}

	return data.Vector, nil
}

// updateToNewFormat updates an existing record from old format to new format
func (s *BadgerStore) updateToNewFormat(id string, data *vectorData) error {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(data); err != nil {
		return err
	}

	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(id), buf.Bytes())
	})
}

func (s *BadgerStore) GetAllVectors() (map[string][]float32, error) {
	vectors := make(map[string][]float32)

	err := s.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			key := string(item.Key())

			var data vectorData
			err := item.Value(func(v []byte) error {
				// First try to decode as the new vectorData struct
				dec := gob.NewDecoder(bytes.NewReader(v))
				err := dec.Decode(&data)
				if err != nil {
					// If that fails, try to decode as the old []float32 format
					var oldVec []float32
					decOld := gob.NewDecoder(bytes.NewReader(v))
					oldErr := decOld.Decode(&oldVec)
					if oldErr != nil {
						return fmt.Errorf("failed to decode vector %s: %w", key, err)
					}
					// Convert to new format with computed norm
					var norm float32
					for _, val := range oldVec {
						norm += val * val
					}
					norm = float32(math.Sqrt(float64(norm)))
					data = vectorData{
						Vector: oldVec,
						Norm:   norm,
						Meta:   nil,
					}
					// Optionally update the stored record to new format
					_ = s.updateToNewFormat(key, &data) // Don't fail if this fails
				}
				return nil
			})
			if err != nil {
				return err
			}

			vectors[key] = data.Vector
		}
		return nil
	})

	return vectors, err
}

func (s *BadgerStore) Close() error {
	return s.db.Close()
}

// vectorData holds the complete vector information including precomputed norm
type vectorData struct {
	Vector []float32
	Norm   float32
	Meta   map[string]any
}

func (s *BadgerStore) Add(id string, vec []float32, meta map[string]any) error {
	// Precompute norm for faster similarity calculations
	var norm float32
	for _, val := range vec {
		norm += val * val
	}
	norm = float32(math.Sqrt(float64(norm)))

	data := vectorData{
		Vector: vec,
		Norm:   norm,
		Meta:   meta,
	}

	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(data); err != nil {
		return err
	}

	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(id), buf.Bytes())
	})
}

func (s *BadgerStore) Get(id string) ([]float32, float32, map[string]any, error) {
	var data vectorData

	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(id))
		if err != nil {
			return err
		}
		return item.Value(func(v []byte) error {
			// First try to decode as the new vectorData struct
			dec := gob.NewDecoder(bytes.NewReader(v))
			err := dec.Decode(&data)
			if err != nil {
				// If that fails, try to decode as the old []float32 format
				var oldVec []float32
				decOld := gob.NewDecoder(bytes.NewReader(v))
				oldErr := decOld.Decode(&oldVec)
				if oldErr != nil {
					return fmt.Errorf("failed to decode vector data: %w", err)
				}
				// Convert to new format with computed norm
				var norm float32
				for _, val := range oldVec {
					norm += val * val
				}
				norm = float32(math.Sqrt(float64(norm)))
				data = vectorData{
					Vector: oldVec,
					Norm:   norm,
					Meta:   nil,
				}
				// Optionally update the stored record to new format
				_ = s.updateToNewFormat(id, &data) // Don't fail if this fails
			}
			return nil
		})
	})

	if err != nil {
		return nil, 0, nil, err
	}

	return data.Vector, data.Norm, data.Meta, nil
}

func (s *BadgerStore) Search(query []float32, k int) ([]embedx.SearchResult, error) {
	results := make([]embedx.SearchResult, 0)

	queryNorm := s.computeNorm(query)

	err := s.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			id := string(item.Key())

			var data vectorData
			err := item.Value(func(v []byte) error {
				// First try to decode as the new vectorData struct
				dec := gob.NewDecoder(bytes.NewReader(v))
				err := dec.Decode(&data)
				if err != nil {
					// If that fails, try to decode as the old []float32 format
					var oldVec []float32
					decOld := gob.NewDecoder(bytes.NewReader(v))
					oldErr := decOld.Decode(&oldVec)
					if oldErr != nil {
						return fmt.Errorf("failed to decode vector %s: %w", id, err)
					}
					// Convert to new format with computed norm
					var norm float32
					for _, val := range oldVec {
						norm += val * val
					}
					norm = float32(math.Sqrt(float64(norm)))
					data = vectorData{
						Vector: oldVec,
						Norm:   norm,
						Meta:   nil,
					}
					// Optionally update the stored record to new format
					_ = s.updateToNewFormat(id, &data) // Don't fail if this fails
				}
				return nil
			})
			if err != nil {
				return err
			}

			if len(data.Vector) != len(query) {
				continue
			}

			// Calculate cosine similarity using precomputed norm
			var dotProduct float32
			for i := range query {
				dotProduct += query[i] * data.Vector[i]
			}

			if queryNorm == 0 || data.Norm == 0 {
				continue
			}

			score := dotProduct / (queryNorm * data.Norm)

			results = append(results, embedx.SearchResult{
				ID:    id,
				Score: score,
				Meta:  data.Meta,
			})
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	// Sort by score descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	// Return top-k results
	if k > 0 && len(results) > k {
		results = results[:k]
	}

	return results, nil
}

// ImportVectors imports vectors from a map
func (s *BadgerStore) ImportVectors(vectors map[string][]float32) error {
	for id, vec := range vectors {
		if err := s.SaveVector(id, vec); err != nil {
			return fmt.Errorf("failed to import vector %s: %w", id, err)
		}
	}
	return nil
}

// ExportVectors exports all vectors to a map
func (s *BadgerStore) ExportVectors() (map[string][]float32, error) {
	return s.GetAllVectors()
}

// computeNorm computes the L2 norm of a vector
func (s *BadgerStore) computeNorm(vec []float32) float32 {
	var norm float32
	for _, val := range vec {
		norm += val * val
	}
	return float32(math.Sqrt(float64(norm)))
}
