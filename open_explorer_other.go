//go:build !windows

package main

import (
	"os/exec"
	"runtime"
)

func openFileExplorer(dir string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", dir).Start()
	default:
		return exec.Command("xdg-open", dir).Start()
	}
}
