//go:build !windows

package pyenv

import "os/exec"

// hideWindow is a no-op on non-Windows platforms.
func hideWindow(_ *exec.Cmd) {}
