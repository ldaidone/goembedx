package vector

import "github.com/ldaidone/goembedx/vector/internal"

// dotGeneric computes the dot product using a generic implementation.
// This is a simple loop-based implementation that works on all architectures.
// It serves as a fallback when optimized implementations are not available.
func dotGeneric(a, b []float32) float32 { return internal.DotGeneric(a, b) }

// dotBlocked computes the dot product using a blocked/unrolled kernel implementation.
// This implementation improves cache efficiency and performance for larger vectors
// by processing elements in blocks of size cfg.BlockSize.
func dotBlocked(a, b []float32, cfg DotConfig) float32 {
	return internal.DotBlocked(a, b, cfg.BlockSize)
}
