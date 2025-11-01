package main

import (
	"fmt"
	"github.com/ldaidone/goembedx/legacy/search"
	"github.com/ldaidone/goembedx/legacy/store"

	"github.com/ldaidone/goembedx/vector"
)

func main() {
	// small demo showing add + search
	s := store.NewMemoryStore(3)
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
}
