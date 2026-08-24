# goembedx Roadmap

This roadmap outlines the planned evolution from minimal engine to production-ready semantic vector database.

---

## ✅ v0.1.0 — Foundation (released)
- In-memory vector store
- Cosine similarity search (brute-force)
- Norm precomputation for speed
- CI + coverage + examples

---

## ✅ v0.2.0 — Persistent Store
- BadgerDB backend
- SQLite backend (multi-process safe)
- Import/export vectors
- CLI: `goembedx add/search`
- Persist vector norms

---

## ✅ v0.3.0 — Performance Upgrade
- Blocked dot product optimization
- DotBatch serial/parallel with auto-tuned threshold
- `scripts/ascii-banner` — portable 3D-shadow banner (rainbow/solid/per-letter color, `NO_COLOR`, `make install-banner`)

---

## 🎯 v0.4.0 — RAG + Search (current)
- `Embedder` interface + deterministic `Dummy` embedder (FNV-hash)
- `Searcher` interface with composable Filter DSL (`Eq`/`In`/`Exists`/`And`/`Or`/`Not`)
- `rag` pipeline: `Retriever` + `FormatContext` + `BuildPrompt` with token-budget management
- `models/get-small.gtemodel` binary asset

---

## 📦 v0.5.0 — Text + Metadata API
- Optional text-embedding helpers (Ollama/OpenAI)
- Payload & metadata store
- Friendly RAG utility functions

---

## ⚙️ v0.6.0 — ANN Prototype
- HNSW graph index (experimental)
- On-disk structure support
- Benchmark suite

---

## 💡 Future
- Prometheus metrics
- LRU + caching layer (gomemo integration)
- gRPC API
- WASM support
- Mobile / edge mode
- SIMD acceleration (AVX2 / NEON) — postponed
- Benchmarks vs Faiss — postponed

---

Community feedback will adapt this roadmap — PRs welcome!
