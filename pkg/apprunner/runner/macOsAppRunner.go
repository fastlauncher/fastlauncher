package runner

import (
	"os/exec"
	"syscall"
)

type macOsAppRunner struct{}

func GetMacOsAppRunner() macOsAppRunner {
	f := macOsAppRunner{}

	return f
}

func (lr *macOsAppRunner) Run(command string) error {
	cmd := exec.Command("bash", "-c", command)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	err := cmd.Start()

	return err
}
