# goembedx 🧠⚡
> Lightweight local embedding store for Go — pure Go, no external services, blazing fast nearest-vector search.

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go Reference](https://pkg.go.dev/badge/github.com/ldaidone/goembedx.svg)](https://pkg.go.dev/github.com/ldaidone/goembedx)
[![Build](https://github.com/ldaidone/goembedx/actions/workflows/ci.yml/badge.svg)](https://github.com/ldaidone/goembedx/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/ldaidone/goembedx/branch/main/graph/badge.svg)](https://codecov.io/gh/ldaidone/goembedx)
![Stars](https://img.shields.io/github/stars/ldaidone/goembedx?style=social)


> 💡 **goembedx** is a tiny vector database for embeddings — perfect for local LLM agents, RAG systems, and semantic search inside Go applications.

### ✨ Features

- 🔥 Pure Go (no CGO, no external libraries)
- ⚡ Fast cosine similarity search
- 📦 In-memory, SQLite, and BadgerDB backends behind one `store.Store` interface
- 🧩 Searcher interface + composable Filter DSL (`Eq`/`In`/`Exists`/`And`/`Or`/`Not`)
- 🧬 Precomputed vector norms for optimized search
- 📁 Import/export vector functionality
- 🧪 Blocked dot products with auto-tuned block size
- ⚡ DotBatch serial/parallel with parallel threshold heuristic
- 🧠 **RAG pipeline** — `Retriever` + `FormatContext` + `BuildPrompt` with token-budget management
- 🧩 `Embedder` interface + deterministic `Dummy` embedder for offline testing
- 🖥️ CLI tools (`goembedx add/search`) for vector management
- 💾 Works offline — great for agents on the edge
- 🧪 Fully tested, clean API, blazing performance
- 🔌 Future: goembedx serve — REST API mode
- ⚠️ Future: SIMD backends (AVX2 / NEON) and Faiss comparison
- 🧠 Future: Optional ANN index (HNSW lite)

---

### 🚀 Quick Start

Every vector store implements the `store.Store` interface
(`SaveVector`/`GetVector`/`GetAllVectors`/`Add`/`Get`/`Search`/`Close`), so
swapping backends is a one-line change. The simplest one is in-memory:

```go
import (
	"fmt"

	"github.com/ldaidone/goembedx/pkg/store"
)

func main() {
	// In-memory store constrained to 384-dim vectors (MiniLM, etc.).
	s := store.NewMemory(384)

	// Add vectors with optional metadata.
	if err := s.Add("doc1", []float32{0.1, 0.2, 0.3}, map[string]any{"title": "hello world"}); err != nil {
		panic(err)
	}
	if err := s.Add("doc2", []float32{0.4, 0.5, 0.6}, map[string]any{"title": "go embedding"}); err != nil {
		panic(err)
	}

	// Cosine similarity search: top-k results sorted by descending score.
	results, err := s.Search([]float32{0.15, 0.25, 0.35}, 3)
	if err != nil {
		panic(err)
	}
	for _, r := range results {
		fmt.Printf("%s %.4f %v\n", r.ID, r.Score, r.Meta)
	}
}
```

### 💾 Persistent storage

Use a BadgerDB- or SQLite-backed store when vectors must survive a restart.
Both expose the same `store.Store` interface — a typical semantic-indexing
workflow reads an existing vector with `Get` (returning the vector, its
precomputed norm, and metadata), skips re-embedding unchanged payloads, and
writes fresh vectors with `Add`:

```go
import "github.com/ldaidone/goembedx/pkg/store"

// BadgerDB-backed store (single process, fastest local access).
badgerStore, err := store.NewBadger("./data/badger")
if err != nil {
	panic(err)
}
defer badgerStore.Close()

// SQLite-backed store (safe for multiple processes sharing the DB file).
sqliteStore, err := store.NewSQLite("./data/sqlite")
if err != nil {
	panic(err)
}
defer sqliteStore.Close()

// Vectors travel with optional metadata, e.g. a payload hash so incremental
// rebuilds only re-embed content that changed since the last run.
const id = "doc1"
const payloadHash = "abc123"

vec, _, meta, err := sqliteStore.Get(id)
if err == nil && meta != nil && meta["payload_hash"] == payloadHash && len(vec) > 0 {
	// Vector already stored for this exact payload — nothing to do.
	return
}

// Embed the (possibly new) payload and store it with its hash.
if err := sqliteStore.Add(id, []float32{0.1, 0.2, 0.3}, map[string]any{"payload_hash": payloadHash}); err != nil {
	panic(err)
}
```

### 🧩 Embed engine & custom stores

The higher-level `embedx` engine needs only the basic vector operations
(`embedx.VectorStore`), which every `store.Store` satisfies:

```go
import "github.com/ldaidone/goembedx/pkg/embedx"

engine := embedx.New(sqliteStore) // *store.SQLite implements embedx.VectorStore
engine.Add("doc3", []float32{0.1, 0.2, 0.3})

results, err := engine.Search([]float32{0.1, 0.2, 0.3}, 3)
```

Because `store.Store` is a public interface, you can also implement your own
backend in your own package and plug it in — no internal packages required.

### 🧩 Search & Filtering

The `Searcher` interface (`embedx.Searcher`) exposes `SearchContext(ctx, query, opts...)`.
Use the Filter DSL to restrict results by metadata before ranking:

```go
import "github.com/ldaidone/goembedx/pkg/embedx"

// Search with metadata filters.
results, err := store.SearchContext(ctx, queryVec,
	embedx.WithK(5),
	embedx.WithFilter(embedx.And(
		embedx.Eq("lang", "en"),
		embedx.Exists("title"),
	)),
)

// Compose filters: Eq, In, Exists, And, Or, Not.
langFilter := embedx.In("lang", "en", "de")
docTypeFilter := embedx.Eq("type", "article")
combined := embedx.And(langFilter, docTypeFilter)
```

### 🧠 RAG Pipeline

The `rag` package ties everything together for Retrieval-Augmented Generation:

```go
import "github.com/ldaidone/goembedx/pkg/rag"

// Create a retriever backed by any Embedder + Searcher.
ret := rag.NewRetriever(myEmbedder, myStore,
	rag.WithTopK(5),       // default top-k
	rag.WithTextKey("text"), // metadata key holding the chunk text
)

// Retrieve relevant chunks for a query.
chunks, err := ret.Retrieve(ctx, "What is Go?")

// Format chunks into a citation-aware context block.
ctxText := rag.FormatContext(chunks, 2048) // token budget

// Build a full RAG prompt for an LLM.
prompt := rag.BuildPrompt("You are a helpful assistant.", "What is Go?", ctxText)
fmt.Println(prompt)
```

### 🖥️ CLI Usage
```bash
# Add a vector with ID
goembedx add doc1 0.1 0.2 0.3 0.4
 
# Search for similar vectors
goembedx search 0.15 0.25 0.35 0.45
```

### 📦 Install

```bash
go get github.com/ldaidone/goembedx
```

### 🔭 Roadmap

Check our complete roadmap and future plans in [ROADMAP.md](./ROADMAP.md).

### 🧪 Testing

```bash
go test ./...
```

or you can use **Makefile** commands

```bash
make test
```

### Makefile help

To know all available commands run

```bash
make help
```

## License

Apache 2.0 License - see the [LICENSE](LICENSE) file for details.

## Support

If this saves you time or helps your AI project, consider starring ⭐
and consider [buying me a coffee](https://www.buymeacoffee.com/leodaido)! ☕️ — it keeps the ideas flowing!

--- 

### ✅ Changes made for v0.3.0:

- RAG pipeline: `Retriever`, `FormatContext`, `BuildPrompt`, token-budget management.
- `Embedder` interface + deterministic `Dummy` embedder (FNV-hash, unit-norm).
- `Searcher` interface with composable Filter DSL (`Eq`/`In`/`Exists`/`And`/`Or`/`Not`).
- Blocked dot products with auto-tuned block size; DotBatch serial/parallel threshold.
- `models/get-small.gtemodel` binary asset.
- `scripts/ascii-banner`: portable 3D-shadow ASCII banner with rainbow/solid/per-letter color, `NO_COLOR` support, and `make install-banner` target.