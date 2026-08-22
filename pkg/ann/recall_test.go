package ann

import (
	"fmt"
	"math"
	"math/rand"
	"testing"

	"github.com/ldaidone/goembedx/vector"
)

const (
	recallDim     = 64
	recallN       = 2000
	recallQueries = 50
	recallK       = 10
)

func normalize(v []float32) {
	n := vector.Norm(v)
	if n == 0 {
		return
	}
	for i := range v {
		v[i] /= n
	}
}

func clusteredDataset(rng *rand.Rand, n int) [][]float32 {
	const clusters = 20
	centers := make([][]float32, clusters)
	for c := range centers {
		center := make([]float32, recallDim)
		for d := range center {
			center[d] = rng.Float32()*2 - 1
		}
		normalize(center)
		centers[c] = center
	}
	vecs := make([][]float32, n)
	for i := range vecs {
		v := make([]float32, recallDim)
		c := rng.Intn(clusters)
		for d := range v {
			v[d] = centers[c][d] + float32(rng.NormFloat64())*0.12
		}
		normalize(v)
		vecs[i] = v
	}
	return vecs
}

func uniformDataset(rng *rand.Rand, n int) [][]float32 {
	vecs := make([][]float32, n)
	for i := range vecs {
		v := make([]float32, recallDim)
		for d := range v {
			v[d] = rng.Float32()*2 - 1
		}
		normalize(v)
		vecs[i] = v
	}
	return vecs
}

func buildIndex(vecs [][]float32, opts ...Option) *HNSW {
	idx := New(opts...)
	for i, v := range vecs {
		if err := idx.Add(fmt.Sprintf("v%d", i), v); err != nil {
			panic(err)
		}
	}
	return idx
}

func exactTopK(vecs [][]float32, q []float32, k int) map[string]bool {
	type pair struct {
		id   string
		dist float64
	}
	all := make([]pair, len(vecs))
	for i, v := range vecs {
		all[i] = pair{id: fmt.Sprintf("v%d", i), dist: cosineDist(q, v)}
	}
	for i := 1; i < len(all); i++ {
		for j := i; j > 0 && all[j].dist < all[j-1].dist; j-- {
			all[j], all[j-1] = all[j-1], all[j]
		}
	}
	if k > len(all) {
		k = len(all)
	}
	got := make(map[string]bool, k)
	for _, p := range all[:k] {
		got[p.id] = true
	}
	return got
}

func cosineDist(a, b []float32) float64 {
	na := float64(vector.Norm(a))
	nb := float64(vector.Norm(b))
	if na == 0 || nb == 0 {
		return 1
	}
	cos := float64(vector.Dot(a, b)) / (na * nb)
	if cos > 1 {
		cos = 1
	} else if cos < -1 {
		cos = -1
	}
	return 1 - cos
}

func measureRecall(idx *HNSW, vecs [][]float32, queries [][]float32, k int) float64 {
	total := 0.0
	for _, q := range queries {
		truth := exactTopK(vecs, q, k)
		hits := idx.Search(q, k)
		matched := 0
		for _, r := range hits {
			if truth[r.ID] {
				matched++
			}
		}
		total += float64(matched) / float64(k)
	}
	return total / float64(len(queries))
}

func TestRecallAtK(t *testing.T) {
	rng := rand.New(rand.NewSource(11))
	clustered := clusteredDataset(rng, recallN)
	uniform := uniformDataset(rng, recallN)
	queries := clusteredDataset(rng, recallQueries)

	cases := []struct {
		name string
		data [][]float32
		opts []Option
		min  float64
	}{
		{"clustered/default", clustered, nil, 0.90},
		{"clustered/ef100", clustered, []Option{WithEfSearch(100)}, 0.98},
		{"uniform/default", uniform, nil, 0.80},
		{"uniform/ef100", uniform, []Option{WithEfSearch(100)}, 0.90},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			idx := buildIndex(tc.data, tc.opts...)
			recall := measureRecall(idx, tc.data, queries, recallK)
			t.Logf("%s: recall@%d = %.3f (n=%d, dim=%d)", tc.name, recallK, recall, len(tc.data), recallDim)
			if recall < tc.min {
				t.Errorf("recall %.3f below floor %.3f", recall, tc.min)
			}
		})
	}
}

func TestRecallSmallM(t *testing.T) {
	rng := rand.New(rand.NewSource(13))
	data := clusteredDataset(rng, recallN)
	queries := clusteredDataset(rng, recallQueries)
	idx := buildIndex(data, WithM(4), WithEfConstruction(80))
	recallDefault := measureRecall(idx, data, queries, recallK)
	t.Logf("M=4 ef=50: recall@%d = %.3f", recallK, recallDefault)

	wide := New(WithM(4), WithEfConstruction(80), WithEfSearch(200))
	for i, v := range data {
		if err := wide.Add(fmt.Sprintf("v%d", i), v); err != nil {
			t.Fatalf("Add failed: %v", err)
		}
	}
	recallWide := measureRecall(wide, data, queries, recallK)
	t.Logf("M=4 ef=200: recall@%d = %.3f", recallK, recallWide)
	if recallWide < 0.85 {
		t.Errorf("recall %.3f below floor 0.85 even at ef=200 (graph connectivity issue)", recallWide)
	}
}

func BenchmarkHNSW_Build_1000(b *testing.B) {
	rng := rand.New(rand.NewSource(21))
	data := clusteredDataset(rng, 1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buildIndex(data)
	}
}

func benchmarkSearch(b *testing.B, ef int) {
	rng := rand.New(rand.NewSource(22))
	data := clusteredDataset(rng, recallN)
	idx := buildIndex(data, WithEfSearch(ef))
	queries := clusteredDataset(rng, 100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		idx.Search(queries[i%len(queries)], recallK)
	}
}

func BenchmarkHNSW_Search_N2000_Ef50(b *testing.B)  { benchmarkSearch(b, 50) }
func BenchmarkHNSW_Search_N2000_Ef100(b *testing.B) { benchmarkSearch(b, 100) }
func BenchmarkHNSW_Search_N2000_Ef200(b *testing.B) { benchmarkSearch(b, 200) }

func BenchmarkBruteForce_N2000(b *testing.B) {
	rng := rand.New(rand.NewSource(22))
	data := clusteredDataset(rng, recallN)
	queries := clusteredDataset(rng, 100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q := queries[i%len(queries)]
		best := math.MaxFloat64
		for _, v := range data {
			if d := cosineDist(q, v); d < best {
				best = d
			}
		}
		_ = best
	}
}
