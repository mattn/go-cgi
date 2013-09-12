package main

import (
	"os/exec"
	"syscall"
)

func command(p string, args ...string) *exec.Cmd {
	cmd := exec.Command(p, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return cmd
}
