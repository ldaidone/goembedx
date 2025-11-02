// go:build amd64
// +build amd64

#include "textflag.h"

// func dotAVX2(a, b []float32) float32
TEXT ·dotAVX2(SB), NOSPLIT, $0-24
    // a in first arg (slice header), b in second arg
    // Layout: [a ptr, a len, a cap] then [b ptr, b len, b cap]
    // We'll load pointers and length, loop over 8 floats (32 bytes) per iter using ymm regs.

    // Load a.data (ptr)  -> RAX
    MOVQ 0(SP), AX
    // Load a.len -> CX
    MOVQ 8(SP), CX
    // Load b.data -> RDX
    MOVQ 24(SP), DX
    // For Plan 9/Go assembler the offsets may need adjustment depending on go version.
    // This template is illustrative. Implementers must validate ABI and adjust.

    // ... real AVX2 asm implementation goes here ...
    // For a safe compile-ready fallback, just call into dotScalar (but that requires CALLing a Go func)
    // We'll simply return 0 here to make this a template.
    XORPS X0, X0
    VMOVSS X0, ret+0(FP)
    RET
