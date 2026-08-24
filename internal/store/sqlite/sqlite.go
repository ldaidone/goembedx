// Package sqlite provides a persistent vector store implementation using SQLite.
// SQLite is an embeddable, file-based relational database written in C but
// driven here through the pure-Go modernc.org/sqlite driver. Unlike BadgerDB,
// SQLite lets several processes open the same database file concurrently, so
// multiple servers can share one store without an exclusive directory lock.
package sqlite

import (
	"bytes"
	"database/sql"
	"encoding/gob"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"

	_ "modernc.org/sqlite" // registers the "sqlite" database driver

	"github.com/ldaidone/goembedx/pkg/embedx"
	"github.com/ldaidone/goembedx/vector"
)

// sqliteDBFileName is the database file created inside the store directory.
const sqliteDBFileName = "vectors.db"

// SQLiteStore implements the embedx stores using SQLite as the persistent backend.
// It stores vectors with precomputed norms for efficient similarity calculations.
type SQLiteStore struct {
	// db is the underlying SQLite database instance.
	db *sql.DB
}

// Compile-time interface checks
var _ embedx.VectorStore = (*SQLiteStore)(nil)
var _ embedx.Store = (*SQLiteStore)(nil)

// NewSQLiteStore creates a new SQLiteStore instance backed by the SQLite
// database file dir/vectors.db. The dir parameter specifies the directory where
// the database file will be stored and is created when it does not exist.
// The connection is configured with WAL journaling, a busy timeout, and NORMAL
// synchronous mode so concurrent processes can read and write safely.
// Returns an error if the database cannot be opened or initialized.
func NewSQLiteStore(dir string) (*SQLiteStore, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create sqlite store directory: %w", err)
	}

	path := filepath.Join(dir, sqliteDBFileName)
	dsn := (&url.URL{Scheme: "file", Path: path}).String() +
		"?_pragma=journal_mode(WAL)&_pragma=busy_timeout(10000)&_pragma=synchronous(NORMAL)"

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}

	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS vectors (id TEXT PRIMARY KEY, data BLOB NOT NULL)`); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to create vectors table: %w", err)
	}

	return &SQLiteStore{db: db}, nil
}

// VectorStore interface methods
func (s *SQLiteStore) SaveVector(id string, vec []float32) error {
	// Use the same data structure as Add to maintain consistency
	data := vectorData{
		Vector: vec,
		Norm:   vector.Norm(vec),
		Meta:   nil, // No metadata for basic SaveVector
	}

	blob, err := encodeVectorData(data)
	if err != nil {
		return err
	}

	return s.upsert(id, blob)
}

func (s *SQLiteStore) GetVector(id string) ([]float32, error) {
	data, err := s.getRaw(id)
	if err != nil {
		return nil, err
	}

	return data.Vector, nil
}

func (s *SQLiteStore) GetAllVectors() (map[string][]float32, error) {
	rows, err := s.db.Query(`SELECT id, data FROM vectors`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	vectors := make(map[string][]float32)
	for rows.Next() {
		var id string
		var blob []byte
		if err := rows.Scan(&id, &blob); err != nil {
			return nil, err
		}
		data, err := decodeVectorData(blob)
		if err != nil {
			return nil, fmt.Errorf("failed to decode vector %s: %w", id, err)
		}
		vectors[id] = data.Vector
	}
	return vectors, rows.Err()
}

func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

// Store interface methods
// Add stores a vector with the given ID and associated metadata.
// It precomputes the L2 norm of the vector for faster similarity calculations.
// Returns an error if the operation fails.
func (s *SQLiteStore) Add(id string, vec []float32, meta map[string]any) error {
	// Precompute norm for faster similarity calculations
	data := vectorData{
		Vector: vec,
		Norm:   vector.Norm(vec),
		Meta:   meta,
	}

	blob, err := encodeVectorData(data)
	if err != nil {
		return err
	}

	return s.upsert(id, blob)
}

// Get retrieves a vector by its ID along with its precomputed norm and metadata.
// Returns the vector, its norm, metadata, and any error that occurred.
func (s *SQLiteStore) Get(id string) ([]float32, float32, map[string]any, error) {
	data, err := s.getRaw(id)
	if err != nil {
		return nil, 0, nil, err
	}

	return data.Vector, data.Norm, data.Meta, nil
}

// Search returns the top-k vectors most similar to the query by cosine
// similarity, computed against the stored precomputed norms. Results are sorted
// by score in descending order.
func (s *SQLiteStore) Search(query []float32, k int) ([]embedx.SearchResult, error) {
	results := make([]embedx.SearchResult, 0)
	queryNorm := vector.Norm(query)

	rows, err := s.db.Query(`SELECT id, data FROM vectors`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id string
		var blob []byte
		if err := rows.Scan(&id, &blob); err != nil {
			return nil, err
		}
		data, err := decodeVectorData(blob)
		if err != nil {
			return nil, fmt.Errorf("failed to decode vector %s: %w", id, err)
		}

		if len(data.Vector) != len(query) {
			continue
		}

		if queryNorm == 0 || data.Norm == 0 {
			continue
		}

		score := vector.Dot(query, data.Vector) / (queryNorm * data.Norm)

		results = append(results, embedx.SearchResult{
			ID:    id,
			Score: score,
			Meta:  data.Meta,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})
	if k > 0 && len(results) > k {
		results = results[:k]
	}
	return results, nil
}

// ImportVectors imports multiple vectors from a map of ID to vector data.
// It stores each vector with its corresponding ID in the SQLite store.
// Returns an error if any vector fails to be imported.
func (s *SQLiteStore) ImportVectors(vectors map[string][]float32) error {
	for id, vec := range vectors {
		if err := s.SaveVector(id, vec); err != nil {
			return fmt.Errorf("failed to import vector %s: %w", id, err)
		}
	}
	return nil
}

// ExportVectors exports all stored vectors to a map of ID to vector data.
// Returns a map of all vectors stored in the database and any error that occurred.
func (s *SQLiteStore) ExportVectors() (map[string][]float32, error) {
	return s.GetAllVectors()
}

// DeleteStale removes every stored vector whose key is not present in valid,
// returning how many were deleted. Used during graph rebuilds so vectors for
// removed files or entities stop surfacing in semantic search.
func (s *SQLiteStore) DeleteStale(valid map[string]struct{}) (int, error) {
	rows, err := s.db.Query(`SELECT id FROM vectors`)
	if err != nil {
		return 0, err
	}

	var stale []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			_ = rows.Close()
			return 0, err
		}
		if _, ok := valid[id]; !ok {
			stale = append(stale, id)
		}
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return 0, err
	}
	_ = rows.Close()

	for _, id := range stale {
		if _, err := s.db.Exec(`DELETE FROM vectors WHERE id = ?`, id); err != nil {
			return 0, err
		}
	}
	return len(stale), nil
}

// vectorData holds the complete vector information including precomputed norm.
// This structure allows efficient retrieval of vectors with their associated metadata and precomputed norms.
type vectorData struct {
	// Vector contains the actual float32 vector data.
	Vector []float32
	// Norm is the precomputed L2 norm of the vector for efficient similarity calculations.
	Norm float32
	// Meta contains optional metadata associated with the vector.
	Meta map[string]any
}

// upsert inserts the gob-encoded record under id, overwriting any existing one.
func (s *SQLiteStore) upsert(id string, blob []byte) error {
	_, err := s.db.Exec(
		`INSERT INTO vectors(id, data) VALUES(?, ?) ON CONFLICT(id) DO UPDATE SET data = excluded.data`,
		id, blob,
	)
	return err
}

// getRaw decodes the record stored under id, transparently migrating records
// written in the legacy []float32-only format.
func (s *SQLiteStore) getRaw(id string) (vectorData, error) {
	var blob []byte
	if err := s.db.QueryRow(`SELECT data FROM vectors WHERE id = ?`, id).Scan(&blob); err != nil {
		return vectorData{}, err
	}
	return decodeVectorData(blob)
}

// encodeVectorData gob-encodes a vector record for storage.
func encodeVectorData(data vectorData) ([]byte, error) {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// decodeVectorData decodes a stored blob, transparently migrating records
// written in the legacy []float32-only format (no precomputed norm).
func decodeVectorData(blob []byte) (vectorData, error) {
	var data vectorData
	dec := gob.NewDecoder(bytes.NewReader(blob))
	if err := dec.Decode(&data); err == nil {
		return data, nil
	}

	var legacy []float32
	decOld := gob.NewDecoder(bytes.NewReader(blob))
	if oldErr := decOld.Decode(&legacy); oldErr != nil {
		return vectorData{}, fmt.Errorf("failed to decode vector data: %w", oldErr)
	}
	return vectorData{Vector: legacy, Norm: vector.Norm(legacy), Meta: nil}, nil
}
