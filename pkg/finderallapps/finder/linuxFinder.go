package finder

import "errors"

type linuxFinder struct{}

func GetLinuxFinder() linuxFinder {
	f := linuxFinder{}

	return f
}

func (lf *linuxFinder) GetAllApp() ([]string, error) {
	// TODO:

	return []string{}, errors.New("Linux is not suport")
}
