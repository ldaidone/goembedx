package vector

import "runtime"

// hasAVX2 checks if the current CPU architecture supports AVX2 instructions.
// Currently, this returns true for x86-64 architecture, though actual AVX2 support
// is determined in cpu_x86.go and cpu_other.go files.
func hasAVX2() bool {
	return runtime.GOARCH == "amd64"
}

// hasNEON checks if the current CPU architecture supports NEON instructions.
// This returns true for ARM64 architecture which typically includes NEON SIMD capabilities.
func hasNEON() bool {
	return runtime.GOARCH == "arm64"
}
