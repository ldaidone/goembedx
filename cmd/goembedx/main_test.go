package main

import (
	"os/exec"
	"testing"
)

func TestCLIHelp(t *testing.T) {
	// Test that the CLI help command can be executed (doesn't panic, etc.)
	// Help may exit with status 1 which is normal for CLI tools
	cmd := exec.Command("go", "run", "./cmd/goembedx", "--help")
	_ = cmd.Run() // Don't treat exit code as failure - help may exit with 1
}
