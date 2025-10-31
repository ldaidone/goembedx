# goembedx Roadmap

This roadmap outlines the planned evolution from minimal engine to production-ready semantic vector database.

---

## ✅ v0.1.0 — Foundation (released)
- In-memory vector store
- Cosine similarity search (brute-force)
- Norm precomputation for speed
- CI + coverage + examples

---

## 🚧 v0.2.0 — Persistent Store (next)
- SQLite backend
- Import/export vectors
- CLI: `goembedx add/search`
- Persist vector norms

---

## 🎯 v0.3.0 — Performance Upgrade
- Blocked dot product optimization
- SIMD acceleration (AVX2 / NEON)
- CPU auto-detect
- Benchmarks vs Faiss-brute

---

## 📦 v0.4.0 — Text + Metadata API
- Optional text-embedding helpers (Ollama/OpenAI)
- Payload & metadata store
- Filtering
- Friendly RAG utility functions

---

## ⚙️ v0.5.0 — ANN Prototype
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

---

Community feedback will adapt this roadmap — PRs welcome!
