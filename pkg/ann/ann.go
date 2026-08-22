// Package ann provides a pure-Go, CGo-free approximate nearest neighbor
// index. It implements the Hierarchical Navigable Small World (HNSW) graph
// algorithm (Malkov & Yashunin, 2016), storing vectors in memory and ranking
// query results by cosine similarity.
//
// Results are approximate by design: a higher EfSearch trades recall for
// speed. Index construction and search are deterministic for a given seed,
// so tests and demos are reproducible.
package ann

import (
	"container/heap"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"sort"

	"github.com/ldaidone/goembedx/vector"
)

// Index is the approximate nearest neighbor index interface.
type Index interface {
	// Add inserts a vector under id. All vectors must share the same
	// dimension, set by the first Add.
	Add(id string, vec []float32) error
	// Search returns the k approximate nearest neighbors to query, ordered
	// by descending cosine similarity.
	Search(query []float32, k int) []Result
	// Len returns the number of stored vectors.
	Len() int
}

// Result is a single approximate nearest neighbor hit.
type Result struct {
	ID    string
	Score float32
}

// ErrDimensionMismatch is returned by Add when the vector dimension differs
// from the dimension of previously stored vectors.
var ErrDimensionMismatch = errors.New("ann: dimension mismatch")

// Config holds HNSW construction and search parameters.
type Config struct {
	// M is the maximum number of outgoing connections per node per layer
	// (except layer 0, which allows 2*M). Defaults to 16.
	M int
	// EfConstruction controls recall during index construction (larger is
	// more accurate, slower to build). Defaults to 200.
	EfConstruction int
	// EfSearch controls recall during Search (larger is more accurate,
	// slower). Defaults to 50.
	EfSearch int
	// Seed makes level assignment and tie-breaking reproducible.
	Seed int64
}

// DefaultConfig returns a Config with the standard HNSW defaults.
func DefaultConfig() Config {
	return Config{M: 16, EfConstruction: 200, EfSearch: 50, Seed: 42}
}

// Option mutates a Config.
type Option func(*Config)

// WithM sets the maximum out-degree per layer.
func WithM(m int) Option {
	return func(c *Config) { c.M = m }
}

// WithEfConstruction sets the construction beam width.
func WithEfConstruction(ef int) Option {
	return func(c *Config) { c.EfConstruction = ef }
}

// WithEfSearch sets the search beam width.
func WithEfSearch(ef int) Option {
	return func(c *Config) { c.EfSearch = ef }
}

// WithSeed sets the random seed.
func WithSeed(seed int64) Option {
	return func(c *Config) { c.Seed = seed }
}

// NewConfig applies opts on top of DefaultConfig.
func NewConfig(opts ...Option) Config {
	cfg := DefaultConfig()
	for _, o := range opts {
		o(&cfg)
	}
	return cfg
}

type hnswNode struct {
	id        string
	vec       []float32
	norm      float64
	level     int
	neighbors [][]int // neighbors[l] are node indexes reachable at layer l
}

// HNSW is a hierarchical navigable small world index. It satisfies Index.
type HNSW struct {
	cfg      Config
	ml       float64 // 1/ln(M): level distribution exponent
	maxM     int     // out-degree cap for layers > 0
	maxM0    int     // out-degree cap for layer 0
	rng      *rand.Rand
	vectors  []hnswNode
	entry    int // index of the top-layer entry point; -1 when empty
	maxLevel int
	dim      int
}

// New returns an empty HNSW index with the given options applied on top of
// the default config.
func New(opts ...Option) *HNSW {
	cfg := NewConfig(opts...)
	if cfg.M < 2 {
		cfg.M = 2
	}
	if cfg.EfConstruction < 1 {
		cfg.EfConstruction = 1
	}
	if cfg.EfSearch < 1 {
		cfg.EfSearch = 1
	}
	return &HNSW{
		cfg:      cfg,
		ml:       1 / math.Log(float64(cfg.M)),
		maxM:     cfg.M,
		maxM0:    2 * cfg.M,
		rng:      rand.New(rand.NewSource(cfg.Seed)),
		entry:    -1,
		maxLevel: 0,
	}
}

// Len returns the number of stored vectors.
func (h *HNSW) Len() int {
	return len(h.vectors)
}

// Add inserts a vector under id, wiring it into every graph layer up to its
// randomly assigned level.
func (h *HNSW) Add(id string, vec []float32) error {
	if len(vec) == 0 {
		return errors.New("ann: cannot add an empty vector")
	}
	if h.dim == 0 {
		h.dim = len(vec)
	} else if len(vec) != h.dim {
		return fmt.Errorf("%w: got %d, want %d", ErrDimensionMismatch, len(vec), h.dim)
	}

	level := h.randomLevel()
	idx := len(h.vectors)
	h.vectors = append(h.vectors, hnswNode{
		id:        id,
		vec:       vec,
		norm:      float64(vector.Norm(vec)),
		level:     level,
		neighbors: make([][]int, level+1),
	})

	if h.entry == -1 {
		h.entry = idx
		h.maxLevel = level
		return nil
	}

	query := vec
	qn := h.vectors[idx].norm
	ep := h.entry
	for layer := h.maxLevel; layer > level; layer-- {
		ep = h.searchLayer(query, qn, ep, 1, layer)[0].idx
	}

	top := level
	if h.maxLevel < top {
		top = h.maxLevel
	}
	for layer := top; layer >= 0; layer-- {
		w := h.searchLayer(query, qn, ep, h.cfg.EfConstruction, layer)
		neighbors := h.selectNeighbors(w, layer)
		h.link(idx, neighbors, layer)
		if len(w) > 0 {
			ep = w[0].idx
		}
	}

	if level > h.maxLevel {
		h.maxLevel = level
		h.entry = idx
	}
	return nil
}

// Search returns up to k approximate nearest neighbors of query, ordered by
// descending cosine similarity.
func (h *HNSW) Search(query []float32, k int) []Result {
	if h.entry == -1 || k <= 0 {
		return []Result{}
	}

	ep := h.entry
	qn := float64(vector.Norm(query))
	for layer := h.maxLevel; layer > 0; layer-- {
		ep = h.searchLayer(query, qn, ep, 1, layer)[0].idx
	}

	ef := h.cfg.EfSearch
	if k > ef {
		ef = k
	}
	w := h.searchLayer(query, qn, ep, ef, 0)
	if k > len(w) {
		k = len(w)
	}

	results := make([]Result, 0, k)
	for i := 0; i < k; i++ {
		results = append(results, Result{ID: h.vectors[w[i].idx].id, Score: 1 - float32(w[i].dist)})
	}
	return results
}

// randomLevel draws a level from the geometric distribution with parameter
// 1 - e^{-1/mL}, so higher levels grow increasingly rare.
func (h *HNSW) randomLevel() int {
	u := h.rng.Float64()
	if u == 0 {
		u = 1e-12
	}
	return int(math.Floor(-math.Log(u) * h.ml))
}

// link adds node idx to each neighbor's adjacency list and each neighbor to
// idx's list at layer, applying the out-degree cap (with heuristic shrinking)
// where needed.
func (h *HNSW) link(idx int, neighbors []int, layer int) {
	n := &h.vectors[idx]
	n.neighbors[layer] = append(n.neighbors[layer], neighbors...)

	cap := h.maxM
	if layer == 0 {
		cap = h.maxM0
	}
	for _, nb := range neighbors {
		nn := &h.vectors[nb]
		nn.neighbors[layer] = append(nn.neighbors[layer], idx)
		if len(nn.neighbors[layer]) > cap {
			h.shrink(nb, layer, cap)
		}
	}
}

// shrink rebuilds a node's adjacency list at layer, keeping the cap closest
// connections by the selection heuristic.
func (h *HNSW) shrink(idx int, layer int, cap int) {
	node := &h.vectors[idx]
	w := make([]candidate, 0, len(node.neighbors[layer]))
	for _, nb := range node.neighbors[layer] {
		w = append(w, candidate{idx: nb, dist: h.distNorm(node.vec, node.norm, h.vectors[nb].vec, h.vectors[nb].norm)})
	}
	// sort ascending by distance to node
	sortCandidates(w)
	kept := h.selectNeighbors(w, layer)
	node.neighbors[layer] = kept
}

// selectNeighbors applies the HNSW heuristic: a candidate is connected only
// if it is closer to the query point than to any already-selected neighbor
// (pruned candidates backfill if the target count is not reached).
func (h *HNSW) selectNeighbors(w []candidate, layer int) []int {
	target := h.maxM
	if layer == 0 {
		target = h.maxM0
	}
	selected := make([]int, 0, target)
	pruned := make([]int, 0, len(w))
	for _, c := range w {
		if len(selected) >= target {
			break
		}
		dominated := false
		for _, s := range selected {
			sn := &h.vectors[s]
			if h.distNorm(h.vectors[c.idx].vec, h.vectors[c.idx].norm, sn.vec, sn.norm) < c.dist {
				dominated = true
				break
			}
		}
		if dominated {
			pruned = append(pruned, c.idx)
		} else {
			selected = append(selected, c.idx)
		}
	}
	for len(selected) < target && len(pruned) > 0 {
		selected = append(selected, pruned[0])
		pruned = pruned[1:]
	}
	return selected
}

// searchLayer runs the beam search of width ef around the query at a single
// graph layer and returns the visited nodes ordered by ascending distance.
func (h *HNSW) searchLayer(query []float32, qn float64, ep int, ef int, layer int) []candidate {
	visited := make(map[int]struct{}, ef*2)
	visited[ep] = struct{}{}

	cand := &candidateHeap{}
	res := &candidateHeap{max: true}
	d0 := h.distNorm(query, qn, h.vectors[ep].vec, h.vectors[ep].norm)
	heap.Push(cand, candidate{idx: ep, dist: d0})
	heap.Push(res, candidate{idx: ep, dist: d0})

	for cand.Len() > 0 {
		c := heap.Pop(cand).(candidate)
		if res.Len() >= ef && c.dist > res.Peek().dist {
			break
		}
		for _, nb := range h.vectors[c.idx].neighbors[layer] {
			if _, seen := visited[nb]; seen {
				continue
			}
			visited[nb] = struct{}{}
			nbNode := &h.vectors[nb]
			d := h.distNorm(query, qn, nbNode.vec, nbNode.norm)
			if res.Len() < ef || d < res.Peek().dist {
				heap.Push(cand, candidate{idx: nb, dist: d})
				heap.Push(res, candidate{idx: nb, dist: d})
				if res.Len() > ef {
					heap.Pop(res)
				}
			}
		}
	}

	out := make([]candidate, 0, res.Len())
	for res.Len() > 0 {
		out = append(out, heap.Pop(res).(candidate))
	}
	// out is descending; reverse to ascending distance.
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return out
}

// distNorm returns the cosine distance between two vectors with precomputed
// norms, in [0, 2].
func (h *HNSW) distNorm(a []float32, na float64, b []float32, nb float64) float64 {
	if na == 0 || nb == 0 {
		return 1
	}
	cos := float64(vector.Dot(a, b)) / (na * nb)
	if cos > 1 {
		cos = 1
	} else if cos < -1 {
		cos = -1
	}
	return 1 - cos
}

// candidate is a node index paired with its distance from a query point.
type candidate struct {
	idx  int
	dist float64
}

// candidateHeap is a binary heap over candidates ordered by distance.
// By default the closest candidate is at the top (min-heap); when max is set,
// the farthest candidate is at the top (max-heap).
type candidateHeap struct {
	max bool
	es  []candidate
}

func (h *candidateHeap) Len() int { return len(h.es) }
func (h *candidateHeap) Less(i, j int) bool {
	if h.max {
		return h.es[i].dist > h.es[j].dist
	}
	return h.es[i].dist < h.es[j].dist
}
func (h *candidateHeap) Swap(i, j int) { h.es[i], h.es[j] = h.es[j], h.es[i] }
func (h *candidateHeap) Push(x any)    { h.es = append(h.es, x.(candidate)) }
func (h *candidateHeap) Pop() any {
	old := h.es
	n := len(old)
	item := old[n-1]
	h.es = old[:n-1]
	return item
}

// Peek returns the top candidate without removing it.
func (h *candidateHeap) Peek() candidate {
	return h.es[0]
}

// sortCandidates sorts candidates by ascending distance in place.
func sortCandidates(w []candidate) {
	sort.Slice(w, func(i, j int) bool { return w[i].dist < w[j].dist })
}
