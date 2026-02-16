package main

import "os/exec"

func openFileExplorer(dir string) error {
	cmd := exec.Command("explorer", dir)
	return cmd.Start()
}
