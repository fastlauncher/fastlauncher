package finderallapps

import (
	"errors"

	"github.com/probeldev/fastlauncher/pkg/finderallapps/finder"
	"github.com/probeldev/fastlauncher/pkg/finderallapps/model"
)

const (
	OsLinux   = "Linux"
	OsMacOs   = "MacOs"
	OsWindows = "Windows"
)

type FinderInterface interface {
	GetAllApp() ([]model.App, error)
}

func GetFinder(operatingSystem string) (FinderInterface, error) {

	switch operatingSystem {
	case OsLinux:
		linuxFinder := finder.GetLinuxFinder()
		return &linuxFinder, nil
	case OsMacOs:
		macOsFinder := finder.GetMacOsFinder()
		return &macOsFinder, nil
	case OsWindows:
		windowsFinder := finder.GetWindowsFinder()
		return &windowsFinder, nil
	}

	return nil, errors.New("operating system not support")
}
