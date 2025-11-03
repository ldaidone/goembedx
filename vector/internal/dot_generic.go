// Package internal contains internal implementation details for vector operations.
// These functions are not part of the public API and are used by the vector package.
package internal

// DotGeneric computes the dot product using a simple generic implementation.
// It performs a straightforward element-wise multiplication and summation.
// This implementation is used as a fallback on architectures without specific optimizations.
func DotGeneric(a, b []float32) float32 {
	var sum float32
	for i := range a {
		sum += a[i] * b[i]
	}
	return sum
}
