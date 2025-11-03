// Package vector provides optimized mathematical operations for float32 vectors.
// It includes SIMD-optimized implementations for common vector operations like dot products.
package vector

import (
	"sync"
)

var (
	dotImpl func(a, b []float32) float32
	once    sync.Once
)

// initDot initializes the optimal dot product implementation based on CPU capabilities.
// It selects between AVX2, NEON, or generic implementations depending on the hardware.
func initDot() {
	switch {
	case hasAVX2():
		// Real implementation later; for now wrap generic/blocked
		dotImpl = func(a, b []float32) float32 {
			ensureAutoTune()
			return dotBlocked(a, b, DefaultDotConfig)
		}
	case hasNEON():
		dotImpl = func(a, b []float32) float32 {
			ensureAutoTune()
			return dotBlocked(a, b, DefaultDotConfig)
		}
	default:
		dotImpl = func(a, b []float32) float32 {
			ensureAutoTune()
			if len(a) > 512 {
				return dotBlocked(a, b, DefaultDotConfig)
			}
			return dotGeneric(a, b)
		}
	}
}

// Dot computes the dot product of two float32 slices.
// It returns the sum of element-wise products: Σ(a[i] * b[i]) for i = 0 to len(a)-1.
// The function automatically selects the optimal implementation based on CPU capabilities.
func Dot(a, b []float32) float32 {
	once.Do(initDot)
	return dotImpl(a, b)
}
