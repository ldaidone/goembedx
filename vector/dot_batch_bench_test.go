package vector

import (
	"math/rand"
	"runtime"
	"testing"
	"time"
)

func makeDB(n, dim int) [][]float32 {
	db := make([][]float32, n)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < n; i++ {
		v := make([]float32, dim)
		for j := range v {
			v[j] = r.Float32()
		}
		db[i] = v
	}
	return db
}

// Force serial via small dim & small batch trick
func BenchmarkDotBatch_ForcedSerial(b *testing.B) {
	dim := 256
	n := runtime.GOMAXPROCS(0) // ensures serial triggers
	db := makeDB(n, dim)
	q := randVec(dim)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = dotBatchSerial(q, db)
	}
}

// Force parallel by bypassing auto logic and calling internal fn
func BenchmarkDotBatch_ForcedParallel(b *testing.B) {
	dim := 768
	n := 20000
	db := makeDB(n, dim)
	q := randVec(dim)

	workers := runtime.GOMAXPROCS(0)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = dotBatchParallel(q, db, workers)
	}
}

// Test real-world behavior (auto mode)
func BenchmarkDotBatch_Auto(b *testing.B) {
	dim := 768
	n := 20000
	db := makeDB(n, dim)
	q := randVec(dim)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = DotBatch(q, db)
	}
}
