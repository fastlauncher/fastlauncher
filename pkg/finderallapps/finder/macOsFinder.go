package finder

import "errors"

type macOsFinder struct{}

func GetMacOsFinder() macOsFinder {
	f := macOsFinder{}

	return f
}

func (mf *macOsFinder) GetAllApp() ([]string, error) {
	// TODO:

	return []string{}, errors.New("MacOs is not suport")
}
