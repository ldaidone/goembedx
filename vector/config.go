package vector

// DotConfig holds configuration parameters for dot product computation.
// These parameters control how the dot product operations are performed,
// particularly for parallel and blocked implementations.
type DotConfig struct {
	// BlockSize specifies the size of blocks for blocked computation algorithms.
	BlockSize int
	// Workers specifies the number of worker goroutines for parallel operations.
	// If 0, runtime.GOMAXPROCS(0) is used.
	Workers int
	// MinDimForParallel is the minimum dimension required to use parallel computation.
	MinDimForParallel int
	// MinBatchFactor determines when to use parallel computation based on
	// batch size relative to worker count.
	MinBatchFactor int
}

// DefaultDotConfig provides reasonable default values for dot product computation.
// These values have been tuned for general-purpose performance across different vector sizes.
var DefaultDotConfig = DotConfig{
	BlockSize:         64,
	Workers:           0,
	MinDimForParallel: 128,
	MinBatchFactor:    4,
}
