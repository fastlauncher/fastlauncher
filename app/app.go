package app

import (
	"os/exec"
	"syscall"
)

type App struct {
}

func (a *App) Run(command string) error {
	cmd := exec.Command("bash", "-c", command)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	err := cmd.Start()

	return err
}
