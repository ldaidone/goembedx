//go:build amd64

// This file contains CPU feature detection specific to x86-64 architecture.
package vector

import (
	"golang.org/x/sys/cpu"
)

// cpuHasAVX2 checks if the CPU supports AVX2 instruction set.
// AVX2 (Advanced Vector Extensions 2) provides enhanced SIMD capabilities
// that can significantly speed up vector operations.
func cpuHasAVX2() bool {
	return cpu.X86.HasAVX2
}
