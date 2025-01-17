package finder

import (
	"errors"

	"github.com/probeldev/fastlauncher/pkg/finderallapps/model"
)

type macOsFinder struct{}

func GetMacOsFinder() macOsFinder {
	f := macOsFinder{}

	return f
}

func (mf *macOsFinder) GetAllApp() ([]model.App, error) {
	apps := []model.App{}
	// TODO:

	return apps, errors.New("MacOs is not suport")
}
