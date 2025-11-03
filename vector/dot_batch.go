package vector

import (
	"runtime"
	"sync"
)

const (
	// defaultChunkSize specifies the number of vectors to process per worker chunk
	defaultChunkSize = 64
	// minDimForParallel sets the minimum dimension required to use parallel processing
	// Below this threshold, serial processing is typically faster
	minDimForParallel = 512
	// minBatchFactor determines when to use parallel processing based on
	// the ratio of batch size to available workers
	minBatchFactor = 4
)

// DotBatch computes dot products of vector `a` against each row in matrix `B`.
// It automatically chooses between serial and parallel computation based on
// vector dimensions and batch size for optimal performance.
// Returns a slice of dot products where result[i] = a · B[i].
func DotBatch(a []float32, B [][]float32) []float32 {
	ensureAutoTune()

	n := len(B)
	if n == 0 {
		return nil
	}

	dim := len(a)
	workers := runtime.GOMAXPROCS(0)

	cfg := DefaultDotConfig

	if dim < cfg.MinDimForParallel || n < workers*cfg.MinBatchFactor {
		return dotBatchSerial(a, B)
	}

	return dotBatchParallel(a, B, workers)
}

// -----------------------
// Serial Implementation
// -----------------------

// dotBatchSerial computes dot products sequentially for small batches or dimensions.
// This implementation is more efficient for small workloads where parallelization overhead would be significant.
func dotBatchSerial(a []float32, B [][]float32) []float32 {
	res := make([]float32, len(B))
	for i := range B {
		res[i] = dotBlocked(a, B[i], DefaultDotConfig) // <- use DefaultDotConfig
	}
	return res
}

// -----------------------
// Parallel Implementation
// -----------------------

// dotBatchParallel computes dot products in parallel using multiple goroutines.
// It distributes the workload across the specified number of workers for improved performance
// on larger batches where parallelization is beneficial.
func dotBatchParallel(a []float32, B [][]float32, workers int) []float32 {
	res := make([]float32, len(B))
	ch := make(chan int, len(B))

	for i := range B {
		ch <- i
	}
	close(ch)

	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range ch {
				res[i] = dotBlocked(a, B[i], DefaultDotConfig) // <- use DefaultDotConfig
			}
		}()
	}

	wg.Wait()
	return res
}
