package tests

import (
	"github.com/ldaidone/goembedx/vector"
	"testing"
)

func TestCosine(t *testing.T) {
	a := []float32{1, 0}
	b := []float32{1, 0}
	if vec.Cosine(a, b) != 1 {
		t.Fatal("expected 1")
	}
}
