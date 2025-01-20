package runner

import (
	"os/exec"
	"syscall"
)

type linuxAppRunner struct{}

func GetLinuxAppRunner() linuxAppRunner {
	f := linuxAppRunner{}

	return f
}

func (lr *linuxAppRunner) Run(command string) error {
	cmd := exec.Command("bash", "-c", command)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	err := cmd.Start()

	return err
}
