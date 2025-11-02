// go:build arm64
// +build arm64

#include "textflag.h"

// func dotNEON(a, b []float32) float32
TEXT ·dotNEON(SB), NOSPLIT, $0-24
    // ARM64 ASIMD template: load pointers, loop using v0..vN
    // This is a template placeholder.
    // Emit zero and return to be safe.
    MOVD $0, R0
    // Convert integer 0 to float32 0.0 in return slot if needed.
    // Simpler: call dotScalar for now, or just return 0.0
    RET
