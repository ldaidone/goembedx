//go:build amd64

package vector

import (
	"golang.org/x/sys/cpu"
)

func cpuHasAVX2() bool {
	return cpu.X86.HasAVX2
}
