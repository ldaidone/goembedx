package vector

import "runtime"

func hasAVX2() bool {
	return runtime.GOARCH == "amd64" && cpuHasAVX2()
}

func hasNEON() bool {
	return runtime.GOARCH == "arm64"
}
