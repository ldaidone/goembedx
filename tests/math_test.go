package tests

import (
	"github.com/ldaidone/goembedx/vector"
	"testing"
)

func TestDotAndNorm(t *testing.T) {
	a := []float32{1, 2, 3}
	b := []float32{4, -5, 6}

	// manual dot = 1*4 + 2*(-5) + 3*6 = 4 -10 +18 = 12
	if got := vector.Dot(a, b); got != 12 {
		t.Fatalf("Dot expected 12, got %v", got)
	}

	// Norm(a)^2 = 1+4+9 = 14 -> Norm = sqrt(14)
	normA := vector.Norm(a)
	if normA <= 0 {
		t.Fatalf("Norm expected >0, got %v", normA)
	}
}

func TestCosineIdentity(t *testing.T) {
	a := []float32{1, 0, 0}
	if got := vector.Cosine(a, a); got != 1 {
		t.Fatalf("Cosine identity expected 1, got %v", got)
	}
}

func TestCosineOrthogonal(t *testing.T) {
	a := []float32{1, 0, 0}
	b := []float32{0, 1, 0}
	if got := vector.Cosine(a, b); got != 0 {
		t.Fatalf("Cosine orthogonal expected 0, got %v", got)
	}
}

func TestCosineNegative(t *testing.T) {
	a := []float32{1}
	b := []float32{-1}
	if got := vector.Cosine(a, b); got != -1 {
		t.Fatalf("Cosine negative expected -1, got %v", got)
	}
}
