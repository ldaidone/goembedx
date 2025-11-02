package internal

// DotBlocked Blocked dot product: tuned pure-Go implementation.
// Block size tuned for cache; inner loop unrolled.
// Safe: expects len(a) == len(b).
func DotBlocked(a, b []float32) float32 {
	n := len(a)
	if n == 0 {
		return 0
	}

	// Block in elements (not bytes). 128 is a reasonable starting point.
	// Tune for your target CPU (32,64,128,256).
	const block = 128

	var sum float32
	for i := 0; i < n; i += block {
		end := i + block
		if end > n {
			end = n
		}

		// Unroll inner loop by 8 for better instruction-level parallelism
		j := i
		// fast path: j + 7 < end
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
