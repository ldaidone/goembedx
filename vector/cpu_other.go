//go:build !amd64

package vector

func cpuHasAVX2() bool { return false }
