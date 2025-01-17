package finder

import (
	"errors"
	"os"
	"strings"
)

type linuxFinder struct{}

func GetLinuxFinder() linuxFinder {
	f := linuxFinder{}

	return f
}

func (lf *linuxFinder) GetAllApp() ([]string, error) {

	foldersApps := []string{
		"/usr/share/applications/",
	}

	for _, folder := range foldersApps {
		lf.getFromFolder(folder)
	}

	// TODO:

	return []string{}, errors.New("Linux is not suport")
}

func (lf *linuxFinder) getFromFolder(folder string) ([]string, error) {
	// TODO
	return []string{}, nil
}

func (lf *linuxFinder) getAllDesktopListFromFolder(folder string) (
	[]string,
	error,
) {
	desktopFiles := []string{}
	entries, err := os.ReadDir(folder)
	if err != nil {
		return desktopFiles, nil
	}
	for _, e := range entries {
		name := e.Name()

		if strings.HasSuffix(name, ".desktop") {
			desktopFiles = append(desktopFiles, folder+name)
		}
	}

	return []string{}, nil
}
