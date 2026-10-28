//go:build !windows

package runner

import (
	"os/exec"
	"syscall"
)

// setProcessGroup makes timeouts kill the whole process tree (javac/java/go
// spawn children), not only the direct child.
func setProcessGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		if cmd.Process == nil {
			return nil
		}
		return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}
}
