package vector

import (
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"
)

var (
	// AutoBlockSize stores the automatically tuned block size for vector operations.
	AutoBlockSize int
	// onceTune ensures that block size tuning runs only once.
	onceTune sync.Once
)

// blockCandidates contains the candidate block sizes to test during auto-tuning.
// These values represent different chunk sizes for blocked vector operations.
var blockCandidates = []int{16, 32, 64, 128, 256}

// randVec generates a random vector of the specified dimension with random float32 values.
// This function is used for benchmarking during the auto-tuning process.
func randVec(dim int) []float32 {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	v := make([]float32, dim)
	for i := range v {
		v[i] = r.Float32()
	}
	return v
}

// tuneBlockSize performs a small benchmark to determine the optimal block size for vector operations.
// It tests different block sizes on sample vectors and selects the one with the best performance.
// The function respects the GEMBEDX_BLOCK environment variable for manual override.
func tuneBlockSize() int {
	// Manual override via environment variable
	if v := os.Getenv("GEMBEDX_BLOCK"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}

	dim := 256
	vecA := randVec(dim)
	vecB := randVec(dim)

	bestBlock := blockCandidates[0]
	bestTime := time.Duration(1<<63 - 1)

	for _, bs := range blockCandidates {
		start := time.Now()
		// small loop to stabilize measurement
		for i := 0; i < 2000; i++ {
			dotBlocked(vecA, vecB, DotConfig{BlockSize: bs})
		}
		elapsed := time.Since(start)
		if elapsed < bestTime {
			bestTime = elapsed
			bestBlock = bs
		}
	}

	return bestBlock
}

// ensureAutoTune runs the block size tuning process exactly once.
// This function should be called by DotBatch and Dot functions to ensure
// optimal performance parameters are set before computation.
func ensureAutoTune() {
	onceTune.Do(func() {
		DefaultDotConfig.BlockSize = tuneBlockSize()
	})
}
