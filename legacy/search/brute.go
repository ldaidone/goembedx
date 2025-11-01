package search

import (
	"container/heap"
	"github.com/ldaidone/goembedx/legacy/store"
	"sort"

	"github.com/ldaidone/goembedx/vector"
)

// Result is a single search result (ID + score).
type Result struct {
	ID    string
	Score float32
}

// SearchBrute returns the top-k results from the provided MemoryStore for the given query vector.
// If k <= 0 it returns all results sorted by score descending.
// Complexity: O(n log k) using a min-heap of size k.
func SearchBrute(s *store.MemoryStore, query []float32, k int) []Result {
	if s == nil {
		return nil
	}
	n := s.Len()
	if n == 0 {
		return []Result{}
	}

	// Compute query norm once (big win)
	qNorm := vector.Norm(query)

	// if k <= 0 or k >= n, compute all and sort
	if k <= 0 || k >= n {
		out := make([]Result, 0, n)
		for _, v := range s.Data() {
			score := vector.Dot(query, v.Val) / (qNorm * v.Norm)
			out = append(out, Result{ID: v.ID, Score: score})
		}
		sort.Slice(out, func(i, j int) bool { return out[i].Score > out[j].Score })
		return out
	}

	// use a min-heap to maintain top-k
	h := &minHeap{}
	heap.Init(h)
	for _, v := range s.Data() {
		score := vector.Dot(query, v.Val) / (qNorm * v.Norm)
		if h.Len() < k {
			heap.Push(h, &item{res: Result{ID: v.ID, Score: score}})
		} else if score > (*h)[0].res.Score {
			heap.Pop(h)
			heap.Push(h, &item{res: Result{ID: v.ID, Score: score}})
		}
	}

	// pop heap into results (reverse order)
	res := make([]Result, h.Len())
	for i := len(res) - 1; i >= 0; i-- {
		it := heap.Pop(h).(*item)
		res[i] = it.res
	}
	return res
}

// min-heap implementation for top-K
type item struct {
	res Result
}

type minHeap []*item

func (h minHeap) Len() int           { return len(h) }
func (h minHeap) Less(i, j int) bool { return h[i].res.Score < h[j].res.Score } // min-heap
func (h minHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *minHeap) Push(x any) {
	*h = append(*h, x.(*item))
}

func (h *minHeap) Pop() any {
	old := *h
	n := len(old)
	it := old[n-1]
	*h = old[:n-1]
	return it
}
