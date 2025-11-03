package internal

import (
	"fmt"
	"testing"
)

// TestDotGeneric tests the basic dot product implementation
func TestDotGeneric(t *testing.T) {
	tests := []struct {
		name     string
		a        []float32
		b        []float32
		expected float32
	}{
		{
			name:     "basic dot product",
			a:        []float32{1, 2, 3},
			b:        []float32{4, 5, 6},
			expected: 32, // 1*4 + 2*5 + 3*6 = 4 + 10 + 18 = 32
		},
		{
			name:     "zero vectors",
			a:        []float32{0, 0, 0},
			b:        []float32{0, 0, 0},
			expected: 0,
		},
		{
			name:     "zero length vectors",
			a:        []float32{},
			b:        []float32{},
			expected: 0,
		},
		{
			name:     "negative values",
			a:        []float32{-1, 2, -3},
			b:        []float32{4, -5, 6},
			expected: -32, // -1*4 + 2*(-5) + (-3)*6 = -4 -10 -18 = -32
		},
		{
			name:     "single element",
			a:        []float32{5},
			b:        []float32{3},
			expected: 15,
		},
		{
			name:     "fractional values",
			a:        []float32{0.5, 1.5},
			b:        []float32{2.0, 4.0},
			expected: 7.0, // 0.5*2.0 + 1.5*4.0 = 1.0 + 6.0 = 7.0
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DotGeneric(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("DotGeneric(%v, %v) = %f, want %f", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

// TestSetBlockSize tests the block size setting function
func TestSetBlockSize(t *testing.T) {
	// Save original value to restore later
	originalSize := GetBlockSize()

	// Test positive values
	SetBlockSize(256)
	if GetBlockSize() != 256 {
		t.Errorf("SetBlockSize(256) failed, GetBlockSize() = %d, want 256", GetBlockSize())
	}

	// Test zero value (should not change)
	SetBlockSize(0)
	if GetBlockSize() != 256 {
		t.Errorf("SetBlockSize(0) should not change value, GetBlockSize() = %d, want 256", GetBlockSize())
	}

	// Test negative value (should not change)
	SetBlockSize(-10)
	if GetBlockSize() != 256 {
		t.Errorf("SetBlockSize(-10) should not change value, GetBlockSize() = %d, want 256", GetBlockSize())
	}

	// Test positive value again
	SetBlockSize(64)
	if GetBlockSize() != 64 {
		t.Errorf("SetBlockSize(64) failed, GetBlockSize() = %d, want 64", GetBlockSize())
	}

	// Restore original value
	SetBlockSize(originalSize)
}

// TestDotBlocked tests the blocked dot product implementation
func TestDotBlocked(t *testing.T) {
	tests := []struct {
		name     string
		a        []float32
		b        []float32
		block    int
		expected float32
	}{
		{
			name:     "basic dot product with default block",
			a:        []float32{1, 2, 3},
			b:        []float32{4, 5, 6},
			block:    64, // default
			expected: 32, // 1*4 + 2*5 + 3*6 = 32
		},
		{
			name:     "zero vectors",
			a:        []float32{0, 0, 0},
			b:        []float32{0, 0, 0},
			block:    32,
			expected: 0,
		},
		{
			name:     "zero length vectors",
			a:        []float32{},
			b:        []float32{},
			block:    16,
			expected: 0,
		},
		{
			name:     "negative values",
			a:        []float32{-1, 2, -3},
			b:        []float32{4, -5, 6},
			block:    8,
			expected: -32,
		},
		{
			name:     "single element",
			a:        []float32{5},
			b:        []float32{3},
			block:    1,
			expected: 15,
		},
		{
			name:     "large vectors",
			a:        []float32{1, 2, 3, 4, 5, 6, 7, 8},
			b:        []float32{1, 1, 1, 1, 1, 1, 1, 1},
			block:    4,
			expected: 36, // sum of 1-8 = 36
		},
		{
			name:     "block size zero uses default",
			a:        []float32{2, 3},
			b:        []float32{5, 7},
			block:    0,  // should use default of 64
			expected: 31, // 2*5 + 3*7 = 10 + 21 = 31
		},
		{
			name:     "block size negative uses default",
			a:        []float32{2, 3},
			b:        []float32{5, 7},
			block:    -5, // should use default of 64
			expected: 31, // 2*5 + 3*7 = 10 + 21 = 31
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DotBlocked(tt.a, tt.b, tt.block)
			if result != tt.expected {
				t.Errorf("DotBlocked(%v, %v, %d) = %f, want %f", tt.a, tt.b, tt.block, result, tt.expected)
			}
		})
	}
}

// TestDotBlockedBlockSizes tests different block sizes with the same vectors
func TestDotBlockedBlockSizes(t *testing.T) {
	a := []float32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	b := []float32{1, 1, 1, 1, 1, 1, 1, 1, 1, 1}

	// Expected result is sum of a = 55
	expected := float32(55.0)

	// Test different block sizes
	blockSizes := []int{1, 2, 4, 8, 16, 32, 64, 128}

	for _, blockSize := range blockSizes {
		t.Run(fmt.Sprintf("block_size_%d", blockSize), func(t *testing.T) {
			result := DotBlocked(a, b, blockSize)
			if result != expected {
				t.Errorf("DotBlocked with block size %d = %f, want %f", blockSize, result, expected)
			}
		})
	}
}

// TestDotBlockedSpecificBlockSizes tests specific scenarios that exercise the unrolling logic
func TestDotBlockedSpecificBlockSizes(t *testing.T) {
	// Test case to ensure proper unrolling behavior with different remainder calculations
	t.Run("remainder_calculation", func(t *testing.T) {
		// This test ensures that elements beyond the unrolled portion are still processed
		a := []float32{1, 2, 3, 4, 5, 6, 7} // 7 elements
		b := []float32{1, 1, 1, 1, 1, 1, 1} // All 1s for easy sum
		blockSize := 8                      // Larger than our vector, but should still process all elements

		result := DotBlocked(a, b, blockSize)
		expected := float32(28.0) // 1*1 + 2*1 + 3*1 + 4*1 + 5*1 + 6*1 + 7*1 = 28

		if result != expected {
			t.Errorf("DotBlocked with vector size 7 and block size 8 = %f, want %f", result, expected)
		}
	})

	// Test with block size that causes unrolling with remainder
	t.Run("block_with_unrolling", func(t *testing.T) {
		a := []float32{1, 2, 3, 4, 5, 6, 7, 8, 9} // 9 elements
		b := []float32{1, 1, 1, 1, 1, 1, 1, 1, 2} // Last element is 2
		blockSize := 8                            // Block size of 8 - will process first 8 with unrolling, last 1 with normal loop

		result := DotBlocked(a, b, blockSize)
		expected := float32(54.0) // 1*1 + 2*1 + 3*1 + 4*1 + 5*1 + 6*1 + 7*1 + 8*1 + 9*2 = 36 + 18 = 54

		if result != expected {
			t.Errorf("DotBlocked with 9 elements and block 8 = %f, want %f", result, expected)
		}
	})
}

// TestGetBlockSize tests the GetBlockSize function
func TestGetBlockSize(t *testing.T) {
	originalSize := GetBlockSize()

	// Change block size
	SetBlockSize(200)
	if GetBlockSize() != 200 {
		t.Errorf("After SetBlockSize(200), GetBlockSize() = %d, want 200", GetBlockSize())
	}

	// Restore original
	SetBlockSize(originalSize)
	if GetBlockSize() != originalSize {
		t.Errorf("After restoring original, GetBlockSize() = %d, want %d", GetBlockSize(), originalSize)
	}
}
