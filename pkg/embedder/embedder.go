// Package embedder defines the Embedder interface and provides a
// deterministic, dependency-free implementation suitable for offline use,
// testing, and prototyping.
package embedder

import (
	"fmt"
	"hash/fnv"
	"math"
)

// Embedder turns a piece of text into a fixed-dimension embedding vector.
type Embedder interface {
	// Embed returns a unit-norm vector representing text. The returned
	// vector has the same dimension for every input of a given instance.
	Embed(text string) ([]float32, error)
}

// Config holds embedder settings.
type Config struct {
	// Dimension is the size of the embedding vectors produced by the
	// embedder. It must be positive.
	Dimension int
}

// DefaultConfig returns a Config with sensible defaults (384 dimensions,
// matching common sentence-embedding models).
func DefaultConfig() Config {
	return Config{Dimension: 384}
}

// Option mutates a Config.
type Option func(*Config)

// WithDimension sets the embedding dimension.
func WithDimension(d int) Option {
	return func(c *Config) { c.Dimension = d }
}

// NewConfig applies opts on top of DefaultConfig.
func NewConfig(opts ...Option) Config {
	cfg := DefaultConfig()
	for _, o := range opts {
		o(&cfg)
	}
	return cfg
}

// Dummy is a deterministic, content-hash-based Embedder. It needs no model,
// network, or external dependencies, and produces the same vector for the
// same input on every run, which makes tests and demos reproducible.
//
// Similar texts are NOT guaranteed to produce similar vectors; Dummy is a
// stand-in for a real embedding model, not a semantic embedder.
type Dummy struct {
	cfg  Config
	seed uint64
}

// NewDummy returns a Dummy embedder with the given options applied on top of
// the default config.
func NewDummy(opts ...Option) *Dummy {
	return &Dummy{
		cfg:  NewConfig(opts...),
		seed: 0x9e3779b97f4a7c15,
	}
}

// Dimension returns the size of the vectors produced by the embedder.
func (d *Dummy) Dimension() int {
	return d.cfg.Dimension
}

// Embed returns a deterministic unit-norm vector of Dimension length for text.
func (d *Dummy) Embed(text string) ([]float32, error) {
	dim := d.cfg.Dimension
	if dim <= 0 {
		return nil, fmt.Errorf("embedder: dimension must be positive, got %d", dim)
	}

	vec := make([]float32, dim)
	for i := range vec {
		vec[i] = d.component(text, i)
	}
	normalize(vec)
	return vec, nil
}

// component derives the i-th coordinate from a salted hash of the text, so
// that identical inputs collide and distinct inputs spread across the space.
func (d *Dummy) component(text string, i int) float32 {
	h := fnv.New64a()
	//nolint:errcheck // hash.Hash never returns an error on Write.
	fmt.Fprintf(h, "%d|%s|%d", d.seed, text, i)
	bits := h.Sum64()
	return float32(float64(bits)/float64(math.MaxUint64))*2 - 1
}

// normalize scales vec to unit length in place; zero vectors are left as-is.
func normalize(vec []float32) {
	var sum float64
	for _, v := range vec {
		sum += float64(v) * float64(v)
	}
	norm := math.Sqrt(sum)
	if norm == 0 {
		return
	}
	for i := range vec {
		vec[i] /= float32(norm)
	}
}
