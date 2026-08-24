package rag

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"testing"

	"github.com/ldaidone/goembedx/pkg/embedder"
	"github.com/ldaidone/goembedx/pkg/store"
)

const benchDocs = 1000

func newBenchRetriever(b *testing.B) *Retriever {
	b.Helper()
	s := store.NewMemory(64)
	e := embedder.NewDummy(embedder.WithDimension(64))
	rng := rand.New(rand.NewSource(5))
	words := []string{"alpha", "bravo", "charlie", "delta", "echo", "foxtrot", "golf", "hotel", "india", "juliet"}
	for i := 0; i < benchDocs; i++ {
		parts := make([]string, 30)
		for w := range parts {
			parts[w] = words[rng.Intn(len(words))]
		}
		text := strings.Join(parts, " ")
		vec, err := e.Embed(text)
		if err != nil {
			b.Fatalf("Embed failed: %v", err)
		}
		if err := s.Add(fmt.Sprintf("doc%d", i), vec, map[string]any{"text": text}); err != nil {
			b.Fatalf("Add failed: %v", err)
		}
	}
	return NewRetriever(e, s)
}

func BenchmarkRetrieve_TopK5_N1000(b *testing.B) {
	r := newBenchRetriever(b)
	ctx := context.Background()
	query := "alpha bravo charlie delta echo foxtrot golf hotel india juliet"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := r.Retrieve(ctx, query); err != nil {
			b.Fatalf("Retrieve failed: %v", err)
		}
	}
}

func BenchmarkRetrieve_TopK20_N1000(b *testing.B) {
	r := newBenchRetriever(b)
	ctx := context.Background()
	query := "alpha bravo charlie delta echo foxtrot golf hotel india juliet"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := r.Retrieve(ctx, query, r.TopK(20)); err != nil {
			b.Fatalf("Retrieve failed: %v", err)
		}
	}
}

func BenchmarkFormatContext_10x500chars(b *testing.B) {
	chunks := make([]Chunk, 10)
	for i := range chunks {
		chunks[i] = Chunk{ID: fmt.Sprintf("d%d", i), Text: strings.Repeat("lorem ipsum dolor sit amet ", 20)}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FormatContext(chunks, 512)
	}
}

func BenchmarkBuildPrompt(b *testing.B) {
	ctxText := strings.Repeat("context passage ", 100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		BuildPrompt("You are a helpful assistant.", "What is the meaning of this?", ctxText)
	}
}
