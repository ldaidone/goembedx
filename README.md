# goembedx 🧠⚡
> Lightweight local embedding store for Go — pure Go, zero dependencies, blazing fast nearest-vector search.

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go Reference](https://pkg.go.dev/badge/github.com/ldaidone/goembedx.svg)](https://pkg.go.dev/github.com/ldaidone/goembedx)
[![Go Report Card](https://goreportcard.com/badge/github.com/ldaidone/goembedx?refresh=1
)](https://goreportcard.com/report/github.com/ldaidone/goembedx
)
[![Build](https://github.com/ldaidone/goembedx/actions/workflows/ci.yml/badge.svg)](https://github.com/ldaidone/goembedx/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/ldaidone/goembedx/branch/main/graph/badge.svg)](https://codecov.io/gh/ldaidone/goembedx)
![Stars](https://img.shields.io/github/stars/ldaidone/goembedx?style=social)


> 💡 **goembedx** is a tiny vector database for embeddings — perfect for local LLM agents, RAG systems, and semantic search inside Go applications.

### ✨ Features

- 🔥 Pure Go (no CGO, no external libraries)
- ⚡ Fast cosine similarity search
- 📦 In-memory and persistent storage with BadgerDB backend
- 🖥️ Available: CLI tools (goembedx add/search) for vector management
- 🧬 Available: Precomputed vector norms for optimized search
- 📁 Available: Import/export vector functionality
- 🧪 Available: Blocked dot products with auto-tuned block size
- 🖥️ Available: DotBatch supports serial/parallel with parallel threshold heuristic
- 🤖 Future: Ollama & HF embedding helpers
- 💾 Works offline — great for agents on the edge
- 🧪 Fully tested, clean API, blazing performance
- 🧠 Build semantic search in minutes
- 🧠 Future: Optional ANN index (HNSW lite)
- 🔌 Future: goembedx serve — REST API mode
- ⚠️ Future: SIMD backends (AVX2 / NEON) and Faiss comparison

---

### 🚀 Quick Start

```go
import (
 "fmt"
 "github.com/ldaidone/goembedx"
)

func main() {
	store := goembedx.New384() // 384-dim example (MiniLM, etc.)

 	store.Add("doc1", []float32{0.1, 0.2, 0.3, /* ... more values to match dimension */ })
 	store.Add("doc2", []float32{0.4, 0.5, 0.6, /* ... more values to match dimension */ })

 	query := []float32{0.15, 0.25, 0.35, /* ... same dimension as vectors */ }
 	results := store.Search(query, 3)

    for _, r := range results {
 		fmt.Println(r.ID, r.Score)
 	}
}
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

- Added notes in features: blocked dot products with auto-tuned block size, DotBatch parallel threshold.
- Updated “Future” to reflect SIMD / Faiss postponed.

Everything else remains as-is — badges, formatting, and tone preserved.