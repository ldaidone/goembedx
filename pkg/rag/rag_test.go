package rag

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/ldaidone/goembedx/pkg/embedder"
	"github.com/ldaidone/goembedx/pkg/embedx"
	"github.com/ldaidone/goembedx/pkg/store"
)

func newTestRetriever(t *testing.T, opts ...Option) *Retriever {
	t.Helper()
	s := store.NewMemory(32)
	docs := []struct {
		id   string
		text string
		lang string
	}{
		{"d1", "the quick brown fox jumps over the lazy dog", "en"},
		{"d2", "the lazy dog sleeps all day long", "en"},
		{"d3", "der schnelle braune fuchs springt über den faulen hund", "de"},
		{"d4", "le renard brun rapide saute par-dessus le chien paresseux", "fr"},
	}
	for _, d := range docs {
		vec, err := embedder.NewDummy(embedder.WithDimension(32)).Embed(d.text)
		if err != nil {
			t.Fatalf("Embed failed: %v", err)
		}
		if err := s.Add(d.id, vec, map[string]any{"text": d.text, "lang": d.lang}); err != nil {
			t.Fatalf("Add failed: %v", err)
		}
	}
	return NewRetriever(embedder.NewDummy(embedder.WithDimension(32)), s, opts...)
}

func TestRetrieve(t *testing.T) {
	r := newTestRetriever(t)
	chunks, err := r.Retrieve(context.Background(), "a fast fox jumping over a dog")
	if err != nil {
		t.Fatalf("Retrieve failed: %v", err)
	}
	if len(chunks) == 0 {
		t.Fatal("Expected at least one chunk")
	}
	if chunks[0].Text != "the quick brown fox jumps over the lazy dog" {
		t.Errorf("Expected best match d1, got %+v", chunks[0])
	}
	if chunks[0].Score <= 0 {
		t.Errorf("Expected positive score, got %f", chunks[0].Score)
	}
}

func TestRetrieveTopK(t *testing.T) {
	r := newTestRetriever(t)
	chunks, err := r.Retrieve(context.Background(), "fox dog", r.TopK(2))
	if err != nil {
		t.Fatalf("Retrieve failed: %v", err)
	}
	if len(chunks) != 2 {
		t.Errorf("Expected 2 chunks, got %d", len(chunks))
	}
}

func TestRetrieveWithFilter(t *testing.T) {
	r := newTestRetriever(t)
	chunks, err := r.Retrieve(context.Background(), "fox dog", WithFilter(embedx.Eq("lang", "de")))
	if err != nil {
		t.Fatalf("Retrieve failed: %v", err)
	}
	if len(chunks) != 1 || chunks[0].ID != "d3" {
		t.Errorf("Expected only d3, got %+v", chunks)
	}
}

func TestRetrieveCustomTextKey(t *testing.T) {
	s := store.NewMemory(8)
	vec, _ := embedder.NewDummy(embedder.WithDimension(8)).Embed("payload")
	if err := s.Add("x", vec, map[string]any{"body": "the payload text", "text": "ignored"}); err != nil {
		t.Fatalf("Add failed: %v", err)
	}
	r := NewRetriever(embedder.NewDummy(embedder.WithDimension(8)), s, WithTextKey("body"))
	chunks, err := r.Retrieve(context.Background(), "payload")
	if err != nil {
		t.Fatalf("Retrieve failed: %v", err)
	}
	if len(chunks) != 1 || chunks[0].Text != "the payload text" {
		t.Errorf("Expected text from 'body' key, got %+v", chunks)
	}
}

func TestEstimateTokens(t *testing.T) {
	if got := EstimateTokens(""); got != 1 {
		t.Errorf("Expected 1 token for empty string, got %d", got)
	}
	if got := EstimateTokens("hello world"); got < 1 {
		t.Errorf("Expected positive token count, got %d", got)
	}
}

func TestFormatContext(t *testing.T) {
	chunks := []Chunk{
		{ID: "a", Text: "first chunk content here"},
		{ID: "b", Text: "second chunk content here"},
		{ID: "c", Text: "third chunk content here"},
	}
	ctx := FormatContext(chunks, 1000)
	if !strings.Contains(ctx, "[1] first chunk content here") {
		t.Errorf("Expected first chunk in context, got:\n%s", ctx)
	}
	if !strings.Contains(ctx, "[3] third chunk content here") {
		t.Errorf("Expected third chunk in context, got:\n%s", ctx)
	}
}

func TestFormatContextBudget(t *testing.T) {
	chunks := []Chunk{
		{ID: "a", Text: strings.Repeat("x", 2000)},
		{ID: "b", Text: strings.Repeat("y", 2000)},
	}
	ctx := FormatContext(chunks, 20)
	if EstimateTokens(ctx) > 25 {
		t.Errorf("Context exceeded budget: %d tokens", EstimateTokens(ctx))
	}
	if !strings.HasSuffix(ctx, "…") {
		t.Error("Expected truncated context to end with ellipsis")
	}
	if strings.Contains(ctx, "[2]") {
		t.Error("Second chunk should be dropped when over budget")
	}
}

func TestBuildPrompt(t *testing.T) {
	p := BuildPrompt("You are helpful.", "What is X?", "[1] about X")
	if !strings.Contains(p, "You are helpful.") {
		t.Error("Missing system prompt")
	}
	if !strings.Contains(p, "Question: What is X?") {
		t.Error("Missing question")
	}
	if !strings.Contains(p, "[1] about X") {
		t.Error("Missing context")
	}
	if !strings.Contains(p, "Cite relevant sources as [n]") {
		t.Error("Missing citation instruction")
	}
}

func TestBuildPromptEmpty(t *testing.T) {
	if got := BuildPrompt("", "q", ""); !strings.Contains(got, "Question: q") {
		t.Errorf("Expected bare prompt, got %q", got)
	}
}

type failingSearcher struct{}

func (failingSearcher) SearchContext(context.Context, []float32, ...embedx.SearchOption) ([]embedx.SearchResult, error) {
	return nil, errors.New("store down")
}

func TestRetrieveSurfacesSearchError(t *testing.T) {
	r := NewRetriever(embedder.NewDummy(embedder.WithDimension(8)), failingSearcher{})
	_, err := r.Retrieve(context.Background(), "x")
	if err == nil {
		t.Fatal("Expected error from underlying store")
	}
	if !strings.Contains(err.Error(), "store down") {
		t.Errorf("Expected wrapped store error, got %v", err)
	}
}

func TestTruncateTextUnicode(t *testing.T) {
	got := truncateText("héllo wörld – ünïcode ✓✓", 3)
	if !strings.HasSuffix(got, "…") {
		t.Errorf("Expected ellipsis, got %q", got)
	}
	_ = got // no panic on rune slicing
}

func TestChunkShape(t *testing.T) {
	r := newTestRetriever(t, WithTopK(1))
	chunks, err := r.Retrieve(context.Background(), "fox dog")
	if err != nil {
		t.Fatalf("Retrieve failed: %v", err)
	}
	if len(chunks) != 1 {
		t.Fatalf("Expected 1 chunk, got %d", len(chunks))
	}
	c := chunks[0]
	if c.ID == "" {
		t.Error("Expected non-empty ID")
	}
	if _, ok := c.Meta["lang"]; !ok {
		t.Error("Expected metadata preserved")
	}
	if !reflect.DeepEqual(c.Meta["text"], c.Text) {
		t.Error("Expected Text to mirror the text metadata key")
	}
}
