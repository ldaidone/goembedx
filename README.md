# goembedx 🧠⚡
> Lightweight local embedding store for Go — pure Go, zero dependencies, blazing fast nearest-vector search.

[![Go Reference](https://pkg.go.dev/badge/github.com/ldaidone/goembedx.svg)](https://pkg.go.dev/github.com/ldaidone/goembedx)
[![Go Report Card](https://goreportcard.com/badge/github.com/ldaidone/goembedx?refresh=1
)](https://goreportcard.com/badge/github.com/ldaidone/goembedx?refresh=1
)
![Stars](https://img.shields.io/github/stars/ldaidone/goembedx?style=social)
[![License](https://img.shields.io/badge/license-Apache_2.0-blue.svg)](LICENSE)
[![Build](https://github.com/ldaidone/goembedx/actions/workflows/ci.yml/badge.svg)](https://github.com/ldaidone/goembedx/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/ldaidone/goembedx/branch/main/graph/badge.svg)](https://codecov.io/gh/ldaidone/goembedx)


> 💡 **goembedx** is a tiny vector database for embeddings — perfect for local LLM agents, RAG systems, and semantic search inside Go applications.

### ✨ Features
- 🔥 Pure Go (no CGO, no external libraries)
- ⚡ Fast cosine similarity search
- 📦 Simple in-memory storage (disk persistence coming soon)
- 🤖 Drop-in tool for local AI workflows (Ollama / HF / OpenAI)
- 💾 Works offline — great for agents on the edge
- 🧪 Fully tested, clean API, blazing performance
- 🧠 Build semantic search in minutes

---

### 🚀 Quick Start

```go
import (
	"fmt"

	"github.com/ldaidone/goembedx"
)

func main() {
	store := goembedx.New(384) // 384-dim example (MiniLM, etc.)

	store.Add("doc1", []float32{ /* embedding */ })
	store.Add("doc2", []float32{ /* embedding */ })

	query := []float32{ /* embedding */ }
	results := store.Search(query, 3)

	for _, r := range results {
		fmt.Println(r.ID, r.Score)
	}
}
```

### 📦 Install

```bash
go get github.com/ldaidone/goembedx
```

### 🔭 Roadmap

- ✅ In-memory vector store
- ✅ Cosine similarity + Top-K
- 🧩 File-based persistence (.embedx)
- 🧠 Optional ANN index (HNSW lite)
- 🤖 Ollama & HF embedding helpers
- 🔌 goembedx serve — REST API mode

### 🧪 Testing

```bash
go test ./...
```

## License

Apache 2.0 License - see the [LICENSE](LICENSE) file for details.

## Support

If this saves you time or helps your AI project, consider starring ⭐
and consider [buying me a coffee](https://www.buymeacoffee.com/leodaido)! ☕️ — it keeps the ideas flowing!