package legacy

//
//import (
//	"errors"
//
//	"github.com/ldaidone/goembedx/search"
//	"github.com/ldaidone/goembedx/store"
//)
//
//// Store is the top-level interface to the vector storage engine.
//// For now we only expose memory, but keep interface future-proof when sqlite/bolt arrive.
////type Store interface {
////	Add(id string, vec []float32) error
////	Len() int
////	Dim() int
////	Data() []store.Vector
////}
//
//// MemoryStore returns an in-memory vector store.
//// dim = embedding dimensionality (e.g. 384, 512, 768, 1024)
//func MemoryStore(dim int) Store {
//	return store.NewMemoryStore(dim)
//}
//
//// AddVector adds a vector to any store (helper for fluent API).
//func AddVector(s Store, id string, vec []float32) error {
//	return s.Add(id, vec)
//}
//
//// SearchTopK performs brute-force cosine similarity search.
//// k <= 0 means "return all".
//func SearchTopK(s Store, query []float32, k int) ([]search.Result, error) {
//	if s == nil {
//		return nil, errors.New("nil store")
//	}
//	if len(query) != s.Dim() {
//		return nil, errors.New("query dimension mismatch")
//	}
//	return search.SearchBrute(s.(*store.MemoryStore), query, k), nil
//}
