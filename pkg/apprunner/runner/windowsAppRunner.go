package runner

import (
	"os/exec"
	"syscall"
)

type windowsAppRunner struct{}

func GetWindowsAppRunner() windowsAppRunner {
	f := windowsAppRunner{}

	return f
}

func (lr *windowsAppRunner) Run(command string) error {
	cmd := exec.Command("bash", "-c", command)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	err := cmd.Start()

	return err
}
