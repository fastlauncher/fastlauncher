package apprunner

import (
	"errors"

	"github.com/probeldev/fastlauncher/pkg/apprunner/runner"
)

const (
	OsLinux   = "Linux"
	OsMacOs   = "MacOs"
	OsWindows = "Windows"
)

type AppRunnerInterface interface {
	Run(string) error
}

func GetAppRunner(operatingSystem string) (AppRunnerInterface, error) {

	switch operatingSystem {
	case OsLinux:
		linuxAppRunner := runner.GetLinuxAppRunner()
		return &linuxAppRunner, nil
	case OsMacOs:
		macOsAppRunner := runner.GetMacOsAppRunner()
		return &macOsAppRunner, nil
	case OsWindows:
		windowsAppRunner := runner.GetWindowsAppRunner()
		return &windowsAppRunner, nil
	}

	return nil, errors.New("Operating System not suport")
}
