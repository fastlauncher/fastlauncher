package mode

import (
	"github.com/probeldev/fastlauncher/model"
	"github.com/probeldev/fastlauncher/pkg/finderallapps"
)

type OsMode struct{}

func (o *OsMode) GetAll() ([]model.App, error) {
	os := o.getOs()
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

func (o *OsMode) getOs() string {
	// TODO change
	return finderallapps.OsLinux
}
