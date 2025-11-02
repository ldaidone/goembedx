// impl_stub.go in /vector (not /internal)
package vector

import "github.com/ldaidone/goembedx/vector/internal"

// pure Go fallback functions
func dotGeneric(a, b []float32) float32 { return internal.DotGeneric(a, b) }
func dotBlocked(a, b []float32) float32 { return internal.DotBlocked(a, b) }

// these will be replaced when real AVX/NEON assembly added
func dotAVX2(a, b []float32) float32 { return internal.DotBlocked(a, b) }
func dotNEON(a, b []float32) float32 { return internal.DotBlocked(a, b) }
