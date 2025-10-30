package vector

import "math"

// Dot computes the dot product of two vectors.
// It panics if the input lengths differ.
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

// Norm returns the L2 norm (magnitude) of a vector.
func Norm(a []float32) float32 {
	var sum float32
	for i := 0; i < len(a); i++ {
		sum += a[i] * a[i]
	}
	return float32(math.Sqrt(float64(sum)))
}

// Cosine returns the cosine similarity between two vectors.
// Values range from -1.0 to 1.0. Panics on length mismatch or zero-length vectors.
func Cosine(a, b []float32) float32 {
	if len(a) != len(b) {
		panic("vector: Cosine requires vectors of equal length")
	}
	na := Norm(a)
	nb := Norm(b)
	if na == 0 || nb == 0 {
		// avoid division by zero; treat as undefined — panic for now
		panic("vector: Cosine with zero-length vector")
	}
	return Dot(a, b) / (na * nb)
}
