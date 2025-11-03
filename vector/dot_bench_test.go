package vector

import (
	"testing"
)

//func randVec(n int) []float32 {
//	v := make([]float32, n)
//	for i := range v {
//		v[i] = rand.Float32()
//	}
//	return v
//}

func BenchmarkDot1M(b *testing.B) {
	a := randVec(1_000_000)
	c := randVec(1_000_000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Dot(a, c)
	}
}

func BenchmarkBatch1M_100(b *testing.B) {
	a := randVec(1_000_000)

	B := make([][]float32, 100)
	for i := range B {
		B[i] = randVec(1_000_000)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DotBatch(a, B)
	}
}
