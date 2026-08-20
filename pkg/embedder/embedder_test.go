package embedder

import (
	"math"
	"reflect"
	"testing"

	"github.com/ldaidone/goembedx/vector"
)

func TestDefaultConfig(t *testing.T) {
	if got := DefaultConfig().Dimension; got != 384 {
		t.Errorf("Expected default dimension 384, got %d", got)
	}
}

func TestNewConfigOptions(t *testing.T) {
	if got := NewConfig(WithDimension(128)).Dimension; got != 128 {
		t.Errorf("Expected dimension 128, got %d", got)
	}
	if got := NewConfig().Dimension; got != 384 {
		t.Errorf("Expected default dimension 384, got %d", got)
	}
}

func TestDummyEmbedDeterministic(t *testing.T) {
	e := NewDummy(WithDimension(8))
	v1, err := e.Embed("hello world")
	if err != nil {
		t.Fatalf("Embed failed: %v", err)
	}
	v2, err := e.Embed("hello world")
	if err != nil {
		t.Fatalf("Embed failed: %v", err)
	}
	if !reflect.DeepEqual(v1, v2) {
		t.Error("Embed should be deterministic for identical input")
	}
}

func TestDummyEmbedDimension(t *testing.T) {
	e := NewDummy(WithDimension(16))
	v, err := e.Embed("anything")
	if err != nil {
		t.Fatalf("Embed failed: %v", err)
	}
	if len(v) != 16 {
		t.Errorf("Expected 16 dimensions, got %d", len(v))
	}
	if e.Dimension() != 16 {
		t.Errorf("Expected Dimension()=16, got %d", e.Dimension())
	}
}

func TestDummyEmbedUnitNorm(t *testing.T) {
	e := NewDummy(WithDimension(64))
	v, err := e.Embed("some text")
	if err != nil {
		t.Fatalf("Embed failed: %v", err)
	}
	norm := vector.Norm(v)
	if math.Abs(float64(norm)-1) > 1e-5 {
		t.Errorf("Expected unit norm, got %f", norm)
	}
}

func TestDummyEmbedDistinct(t *testing.T) {
	e := NewDummy(WithDimension(64))
	v1, _ := e.Embed("alpha")
	v2, _ := e.Embed("beta")
	if reflect.DeepEqual(v1, v2) {
		t.Error("Different inputs should produce different vectors")
	}
}

func TestDummyEmbedEmptyText(t *testing.T) {
	e := NewDummy(WithDimension(8))
	v, err := e.Embed("")
	if err != nil {
		t.Fatalf("Embed of empty string failed: %v", err)
	}
	if len(v) != 8 {
		t.Errorf("Expected 8 dimensions for empty text, got %d", len(v))
	}
	norm := vector.Norm(v)
	if math.Abs(float64(norm)-1) > 1e-5 {
		t.Errorf("Expected unit norm for empty text, got %f", norm)
	}
}

func TestDummyEmbedZeroDimension(t *testing.T) {
	e := NewDummy(WithDimension(0))
	if _, err := e.Embed("x"); err == nil {
		t.Error("Expected error for zero dimension")
	}
}

func TestDummyStableAcrossInstances(t *testing.T) {
	a := NewDummy(WithDimension(32))
	b := NewDummy(WithDimension(32))
	va, _ := a.Embed("shared text")
	vb, _ := b.Embed("shared text")
	if !reflect.DeepEqual(va, vb) {
		t.Error("Embed should be stable across Dummy instances with the same config")
	}
}
