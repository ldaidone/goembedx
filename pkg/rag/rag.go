// Package rag provides retrieval-augmented generation building blocks: a
// Retriever that turns a query into an embedding, pulls relevant vectors from
// a store via SearchContext, and maps them to text chunks, plus prompt
// assembly with a simple token budget.
package rag

import (
	"context"
	"fmt"
	"strings"

	"github.com/ldaidone/goembedx/pkg/embedder"
	"github.com/ldaidone/goembedx/pkg/embedx"
)

// Chunk is a retrieved piece of context along with its provenance.
type Chunk struct {
	ID    string
	Text  string
	Score float32
	Meta  map[string]any
}

// Retriever combines an Embedder and an embedx.Searcher (any bundled store)
// to fetch the chunks most relevant to a text query.
type Retriever struct {
	embed   embedder.Embedder
	search  embedx.Searcher
	k       int
	textKey string
}

// Option configures a Retriever.
type Option func(*Retriever)

// WithTopK sets the number of chunks retrieved per query. Defaults to 5.
func WithTopK(k int) Option {
	return func(r *Retriever) { r.k = k }
}

// WithTextKey sets the metadata key that holds each vector's text. Defaults
// to "text".
func WithTextKey(key string) Option {
	return func(r *Retriever) { r.textKey = key }
}

// NewRetriever returns a Retriever that embeds queries with e and searches s.
func NewRetriever(e embedder.Embedder, s embedx.Searcher, opts ...Option) *Retriever {
	r := &Retriever{embed: e, search: s, k: 5, textKey: "text"}
	for _, o := range opts {
		o(r)
	}
	return r
}

// RetrieveOption configures a single retrieval call.
type RetrieveOption func(*retrieveConfig)

type retrieveConfig struct {
	k      int
	filter embedx.Filter
}

// WithFilter restricts retrieval to vectors whose metadata matches.
func WithFilter(f embedx.Filter) RetrieveOption {
	return func(c *retrieveConfig) { c.filter = f }
}

// TopK overrides the number of chunks to retrieve for this call.
func (r *Retriever) TopK(k int) RetrieveOption {
	return func(c *retrieveConfig) { c.k = k }
}

// Retrieve embeds query, searches the underlying store, and maps the results
// to chunks. Chunks are ordered by descending relevance.
func (r *Retriever) Retrieve(ctx context.Context, query string, opts ...RetrieveOption) ([]Chunk, error) {
	cfg := retrieveConfig{k: r.k}
	for _, o := range opts {
		o(&cfg)
	}

	vec, err := r.embed.Embed(query)
	if err != nil {
		return nil, fmt.Errorf("rag: embed query: %w", err)
	}

	searchOpts := []embedx.SearchOption{embedx.WithK(cfg.k)}
	if cfg.filter != nil {
		searchOpts = append(searchOpts, embedx.WithFilter(cfg.filter))
	}
	results, err := r.search.SearchContext(ctx, vec, searchOpts...)
	if err != nil {
		return nil, fmt.Errorf("rag: search: %w", err)
	}

	chunks := make([]Chunk, 0, len(results))
	for _, res := range results {
		text, _ := res.Meta[r.textKey].(string)
		chunks = append(chunks, Chunk{
			ID:    res.ID,
			Text:  text,
			Score: res.Score,
			Meta:  res.Meta,
		})
	}
	return chunks, nil
}

// EstimateTokens approximates the token count of a string. The 4-characters-
// per-token heuristic is good enough for context-budget trimming.
func EstimateTokens(s string) int {
	return len([]rune(s))/4 + 1
}

// FormatContext renders chunks as a numbered, citation-friendly context block,
// trimming chunks and dropping later ones as needed to stay within maxTokens.
func FormatContext(chunks []Chunk, maxTokens int) string {
	var b strings.Builder
	budget := maxTokens
	for i, c := range chunks {
		header := fmt.Sprintf("[%d] ", i+1)
		text := strings.TrimSpace(c.Text)
		available := budget - EstimateTokens(header)
		if available <= 0 || text == "" {
			break
		}
		if EstimateTokens(text) > available {
			text = truncateText(text, available)
		}
		b.WriteString(header)
		b.WriteString(text)
		b.WriteString("\n\n")
		budget -= EstimateTokens(header + text)
	}
	return strings.TrimSpace(b.String())
}

// BuildPrompt assembles a RAG prompt: system instructions, the user query,
// and the formatted context, with a reminder to cite sources.
func BuildPrompt(system, query, context string) string {
	var b strings.Builder
	if system != "" {
		b.WriteString(system)
		b.WriteString("\n\n")
	}
	if context != "" {
		b.WriteString("Context:\n")
		b.WriteString(context)
		b.WriteString("\n\n")
	}
	b.WriteString("Question: ")
	b.WriteString(query)
	b.WriteString("\n")
	b.WriteString("Answer the question using only the context above. Cite relevant sources as [n].")
	return b.String()
}

// truncateText cuts s to approximately maxTokens tokens, appending an
// ellipsis, without splitting a multi-byte rune.
func truncateText(s string, maxTokens int) string {
	if maxTokens <= 1 {
		return "…"
	}
	chars := maxTokens * 4
	runes := []rune(s)
	if chars > len(runes) {
		chars = len(runes)
	}
	cut := runes[:chars]
	return strings.TrimRight(string(cut), " ") + "…"
}
