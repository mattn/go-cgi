package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

func command(p string, args ...string) *exec.Cmd {
	ep, err := exec.LookPath(p)
	if err != nil {
		if goroot := os.Getenv("GOROOT"); goroot != "" {
			ep = filepath.Join(goroot, "bin", p)
		} else {
			ep = p
		}
	}
	cmd := exec.Command(ep, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return cmd
}
