// Package internal contains internal implementation details for vector operations.
// These functions are not part of the public API and are used by the vector package.
package internal

import "sync/atomic"

var blockSize32 int32 = 128 // default block size (elements)

// SetBlockSize sets the block size used for blocked dot product computation.
// The block size must be positive, otherwise the function has no effect.
func SetBlockSize(n int) {
	if n <= 0 {
		return
	}
	atomic.StoreInt32(&blockSize32, int32(n))
}

// GetBlockSize returns the current block size used for blocked dot product computation.
// The block size can be tuned at runtime for performance optimization.
func GetBlockSize() int {
	return int(atomic.LoadInt32(&blockSize32))
}

// DotBlocked computes a blocked/unrolled dot product for improved cache efficiency.
// It processes vectors in blocks of the specified size, unrolling the computation
// inside each block to reduce loop overhead and improve performance.
// The function relies on the block size parameter to be tuneable for optimal performance.
func DotBlocked(a, b []float32, block int) float32 {
	if block <= 0 {
		block = 64 // safe default
	}

	n := len(a)
	if n == 0 {
		return 0
	}

	var sum float32
	for i := 0; i < n; i += block {
		end := i + block
		if end > n {
			end = n
		}
		j := i
		// unroll by 8 inside block
		for j+7 < end {
			sum += a[j]*b[j] +
				a[j+1]*b[j+1] +
				a[j+2]*b[j+2] +
				a[j+3]*b[j+3] +
				a[j+4]*b[j+4] +
				a[j+5]*b[j+5] +
				a[j+6]*b[j+6] +
				a[j+7]*b[j+7]
			j += 8
		}
		for ; j < end; j++ {
			sum += a[j] * b[j]
		}
	}
	return sum
}
