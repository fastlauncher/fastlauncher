package finder

import "errors"

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
