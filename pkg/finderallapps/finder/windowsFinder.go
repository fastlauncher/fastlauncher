package finder

import (
	"errors"

	"github.com/probeldev/fastlauncher/pkg/finderallapps/model"
)

type windowsFinder struct{}

func GetWindowsFinder() windowsFinder {
	f := windowsFinder{}

	return f
}

func (lf *windowsFinder) GetAllApp() ([]model.App, error) {
	apps := []model.App{}
	// TODO:

	return apps, errors.New("windows is not support")
}
