// Package vector provides basic mathematical operations for float32 vectors.
package vector

import "math"

// Dot computes the dot product of two vectors.
// The dot product is the sum of the products of the corresponding entries
// of the two sequences of numbers.
//
// It panics if the input vectors have different lengths.
func Dot(a, b []float32) float32 {
	if len(a) != len(b) {
		panic("vector: Dot requires vectors of equal length")
	}
	var sum float32
	// simple scalar loop; optimized variants (SIMD) will replace this later
	for i := 0; i < len(a); i++ {
		sum += a[i] * b[i]
	}
	return sum
}

// Norm returns the L2 norm (Euclidean length) of a vector.
// The L2 norm is calculated as the square root of the sum of the squares of its elements.
func Norm(a []float32) float32 {
	var sum float32
	// Sum the squares of the elements.
	for i := 0; i < len(a); i++ {
		sum += a[i] * a[i]
	}

	// The L2 norm is the square root of the sum of squares.
	// We cast to float64 for math.Sqrt and then back to float32.
	return float32(math.Sqrt(float64(sum)))
}

// Cosine returns the cosine similarity between two vectors.
// Cosine similarity is a measure of similarity between two non-zero vectors
// that measures the cosine of the angle between them. The result ranges from
// -1.0 (exactly opposite) to 1.0 (exactly the same), with 0.0 indicating
// orthogonality or decorrelation.
//
// This function will panic if:
//   - The vectors have different lengths.
//   - Either vector has a magnitude (L2 norm) of zero.
func Cosine(a, b []float32) float32 {
	if len(a) != len(b) {
		panic("vector: Cosine requires vectors of equal length")
	}
	na := Norm(a)
	nb := Norm(b)

	// Cosine similarity is undefined for zero-magnitude vectors, as the angle
	// is not defined and it would result in a division by zero.
	if na == 0 || nb == 0 {
		// avoid division by zero; treat as undefined — panic for now
		panic("vector: Cosine with zero-length vector")
	}

	// The formula for cosine similarity is: (A · B) / (||A|| * ||B||)
	return Dot(a, b) / (na * nb)
}
