//go:build !amd64

// This file contains CPU feature detection for non-x86-64 architectures.
package vector

// cpuHasAVX2 always returns false on non-x86-64 architectures.
// AVX2 instruction set is specific to x86-64 processors.
func cpuHasAVX2() bool { return false }
