package internal

import (
	"os/exec"
	"syscall"
)

func Reap(pid int) {
	_ = syscall.Kill(-pid, syscall.SIGKILL)
}

// Set parent ground and death signal to be sure that nested processes will be closed
func SetFlags(cmd *exec.Cmd) {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.Setpgid = true
	cmd.SysProcAttr.Pdeathsig = syscall.SIGINT
}

const Shell = "/bin/sh"
