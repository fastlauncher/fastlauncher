package finder

import "errors"

type windowsFinder struct{}

func GetWindowsFinder() windowsFinder {
	f := windowsFinder{}

	return f
}

func (lf *windowsFinder) GetAllApp() ([]string, error) {
	// TODO:

	return []string{}, errors.New("Windows is not suport")
}
