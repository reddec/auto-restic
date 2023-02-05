//go:build !linux && !darwin

package internal

import (
	"os/exec"
)

func Reap(pid int) {}

func SetFlags(cmd *exec.Cmd) {}
