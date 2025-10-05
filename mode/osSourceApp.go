package mode

import (
	"errors"
	"runtime"

	"github.com/probeldev/fastlauncher/model"
	"github.com/probeldev/fastlauncher/pkg/finderallapps"
)

type OsMode struct{}

func (o *OsMode) GetAll() ([]model.App, error) {
	os, err := o.getFinderOs()
	if err != nil {
		return nil, err
	}
	finder, err := finderallapps.GetFinder(os)

	if err != nil {
		return nil, err
	}

	osApps, err := finder.GetAllApp()
	if err != nil {
		return nil, err
	}

	apps := []model.App{}

	for _, oa := range osApps {
		apps = append(apps, model.App{
			Title:       oa.Name,
			Description: oa.Description,
			Command:     oa.Command,
			Keywords:    oa.Keywords,
		})
	}

	return apps, nil

}

func (o *OsMode) getFinderOs() (string, error) {
	currentOs := runtime.GOOS
	switch currentOs {
	case "darwin":
		return finderallapps.OsMacOs, nil
	case "linux":
		return finderallapps.OsLinux, nil
	}

	return "", errors.New("OS is not support")
}
