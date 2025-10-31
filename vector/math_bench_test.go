package vector

import (
	"math/rand"
	"testing"
	"time"
)

func makeVec(n int) []float32 {
	v := make([]float32, n)
	for i := 0; i < n; i++ {
		v[i] = rand.Float32()*2 - 1 // random in [-1,1]
	}
	return v
}

func init() {
	// seed deterministic-ish randomness for benchmarks
	rand.Seed(time.Now().UnixNano())
}

func BenchmarkDot768(b *testing.B) {
	const dim = 768
	a := makeVec(dim)
	c := makeVec(dim)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Dot(a, c)
	}
}

func BenchmarkCosine768(b *testing.B) {
	const dim = 768
	a := makeVec(dim)
	c := makeVec(dim)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Cosine(a, c)
	}
}
