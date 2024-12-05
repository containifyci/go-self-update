package systemd

import (
	"os"
	"os/exec"
)

func SystemdRestartHook() error {
	cmd := exec.Command("systemctl", "restart", "go-self-update")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
