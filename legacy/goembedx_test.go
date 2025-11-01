package legacy_test

import (
	"github.com/ldaidone/goembedx/pkg/embedx"
	"testing"
)

func TestTopLevelAPI(t *testing.T) {
	s := embedx.MemoryStore(3)
	if err := s.Add("x", []float32{1, 0, 0}); err != nil {
		t.Fatal(err)
	}

	res, err := embedx.SearchTopK(s, []float32{1, 0, 0}, 1)
	if err != nil {
		t.Fatal(err)
	}

	if len(res) != 1 || res[0].ID != "x" {
		t.Fatalf("expected x, got %#v", res)
	}
}

func TestDimMismatch(t *testing.T) {
	s := embedx.MemoryStore(3)
	_, err := embedx.SearchTopK(s, []float32{1, 0}, 1)
	if err == nil {
		t.Fatal("expected dimension mismatch error")
	}
}
