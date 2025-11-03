# Changelog

## [v0.3.0] - 2025-11-03
### Added
- **Blocked Dot Product Optimization**: `dotBlocked` implementation with configurable block size for high-performance dot product computation in pure Go.
- **DotBatch Auto-Tuner**: `DotBatch` automatically selects the optimal block size at runtime for your CPU.
- **CPU Auto-Detect**: Dispatcher picks AVX2 / NEON (future support) or falls back to blocked/generic implementations.
- **Parallel Threshold Heuristic**: DotBatch dynamically chooses serial vs parallel computation based on vector dimensions and batch size.

### Changed
- Public API remains stable: `Dot()`, `DotBatch()` unchanged.
- Benchmarks updated to reflect auto-tuning and parallel threshold improvements.

### Fixed
- Removed broken ASM/CGO SIMD backends to ensure cross-platform build and test stability.

### Notes
- SIMD backends (AVX2 / NEON), Faiss comparison, and full auto-tune validation postponed for post-v1.0 versions.
- Environment override: set `GEMBEDX_BLOCK` to force a specific block size for reproducible benchmarking.

### Benchmarks (Intel i7-4770HQ, 256-dim vectors)

| Benchmark                       | ns/op      | B/op   | allocs/op |
|---------------------------------|------------|--------|------------|
| DotBatch_ForcedSerial-8          | 1,245      | 32     | 1          |
| DotBatch_ForcedParallel-8        | 5,609,510  | 246,734| 11         |
| DotBatch_Auto-8                  | 5,163,270  | 246,615| 11         |
