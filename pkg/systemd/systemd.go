package systemd

import (
	"os"
	"os/exec"
)

func SystemdRestartHook(name string) func() error {
	return func() error {
		cmd := exec.Command("systemctl", "restart", name)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}
}
