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
	// if err != nil {
	// 	panic(err)
	// }
	// _, err := exec.Command("bash", "-c", command).Output()
	// if err != nil {
	// 	log.Println("BashController:Run", command)
	// 	log.Println("BashController:Run", err)
	// }

	return err
}
