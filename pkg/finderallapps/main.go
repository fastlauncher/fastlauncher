package finderallapps

import (
	"errors"

	"github.com/probeldev/fastlauncher/pkg/finderallapps/finder"
)

const (
	OsLinux   = "Linux"
	OsMacOs   = "MacOs"
	OsWindows = "Windows"
)

type FinderInterface interface {
	GetAllApp() ([]string, error)
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

	return nil, errors.New("Operating System not suport")
}
