package main

import (
	"fmt"
	"github.com/ldaidone/goembedx/internal/legacy/search"
	"github.com/ldaidone/goembedx/internal/store/memory"
	"github.com/ldaidone/goembedx/vector"
)

// Legacy memory-based store example - will be deprecated.
func main() {
	// small demo showing add + search
	s := memory.NewMemoryStore(3)
	_ = s.Add("doc1", []float32{1, 0, 0})
	_ = s.Add("doc2", []float32{0.9, 0.1, 0})
	_ = s.Add("doc3", []float32{0, 1, 0})

	query := []float32{1, 0, 0}
	results := search.SearchBrute(s, query, 2)

	fmt.Println("Top results:")
	for i, r := range results {
		fmt.Printf("%d) id=%s score=%.5f\n", i+1, r.ID, r.Score)
	}
	// show cosine computed directly
	fmt.Println("Cosine(doc1,query) =", vector.Cosine([]float32{1, 0, 0}, query))

	// benchmark
	a := make([]float32, 1_000_000)
	b := make([]float32, 1_000_000)
	for i := range a {
		a[i] = 1
		b[i] = 2
	}

	fmt.Println("\n\n\nDot(a,b) =", vector.Dot(a, b), "\n\n\n")

}
