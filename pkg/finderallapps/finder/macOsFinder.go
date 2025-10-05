// Package finder need for find all apps instaled in OS
package finder

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/probeldev/fastlauncher/pkg/finderallapps/model"
)

type macOsFinder struct{}

type appFromFolder struct {
	name     string
	fullPath string
}

func GetMacOsFinder() macOsFinder {
	f := macOsFinder{}

	return f
}

func (mf *macOsFinder) GetAllApp() ([]model.App, error) {
	apps := []model.App{}

	foldersApps := mf.GetAllAppsFolders()

	for _, folder := range foldersApps {
		appsFromFolder, err := mf.GetFromFolder(folder)
		if err != nil {
			return apps, err
		}

		apps = append(apps, appsFromFolder...)
	}

	return apps, nil
}

func (mf *macOsFinder) GetFromFolder(folder string) ([]model.App, error) {

	appsList, err := mf.getAllAppListFromFolder(folder)
	if err != nil {
		return nil, err
	}

	apps := []model.App{}

	for _, a := range appsList {

		apps = append(apps, model.App{
			Name:        a.name,
			Description: "Run " + a.name,
			Command:     `open -n "` + a.fullPath + `"`,
		})
	}

	return apps, nil
}

func (mf *macOsFinder) getAllAppListFromFolder(folder string) ([]appFromFolder, error) {
	appList := []appFromFolder{}

	err := filepath.WalkDir(folder, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		if d.IsDir() && strings.HasSuffix(d.Name(), ".app") {
			appList = append(appList, appFromFolder{
				name:     d.Name(),
				fullPath: path,
			})
			return filepath.SkipDir
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return appList, nil
}

func (mf *macOsFinder) GetAllAppsFolders() []string {
	folders := []string{}

	foldersDefault := mf.GetDefaultAppFolders()
	folders = append(folders, foldersDefault...)

	return folders
}

func (mf *macOsFinder) GetDefaultAppFolders() []string {
	return []string{
		"/System/Applications/",
		"/System/Library/CoreServices/",
		"/Applications/",
	}
}
