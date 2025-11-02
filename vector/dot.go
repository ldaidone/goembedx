package vector

import (
	"sync"
)

var dotImpl func(a, b []float32) float32
var once sync.Once

func initDot() {
	switch {
	case hasAVX2():
		dotImpl = dotAVX2
	case hasNEON():
		dotImpl = dotNEON
	default:
		// If vector is large, use blocked; else generic
		dotImpl = func(a, b []float32) float32 {
			if len(a) > 512 {
				return dotBlocked(a, b)
			}
			return dotGeneric(a, b)
		}
	}
}

func Dot(a, b []float32) float32 {
	once.Do(initDot)
	return dotImpl(a, b)
}
