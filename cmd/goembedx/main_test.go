package main

import (
	"os/exec"
	"testing"
)

func TestCLIHelp(t *testing.T) {
	cmd := exec.Command("go", "run", "./main.go", "-h")
	if err := cmd.Run(); err != nil {
		t.Fatal("CLI help should execute", err)
	}
}
