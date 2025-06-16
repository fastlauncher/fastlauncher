package finder

import (
	"os"
	"strings"

	"github.com/probeldev/fastlauncher/pkg/finderallapps/model"
	"github.com/probeldev/fastlauncher/pkg/parsedesktopfile"
)

type linuxFinder struct{}

func GetLinuxFinder() linuxFinder {
	f := linuxFinder{}

	return f
}

func (lf *linuxFinder) GetAllApp() ([]model.App, error) {
	apps := []model.App{}

	foldersApps := lf.GetAllAppsFolders()

	for _, folder := range foldersApps {
		appsFromFolder, err := lf.GetFromFolder(folder)
		if err != nil {
			return apps, err
		}

		apps = append(apps, appsFromFolder...)
	}

	return apps, nil
}

func (lf *linuxFinder) GetFromFolder(folder string) ([]model.App, error) {

	files, err := lf.getAllDesktopListFromFolder(folder)
	if err != nil {
		return nil, err
	}

	apps := []model.App{}

	parser := parsedesktopfile.GetParseDesktopFile()
	for _, file := range files {
		desktop, err := parser.ParseFromFile(file)
		if err != nil {
			return apps, err
		}

		if desktop.Name == "" || desktop.Exec == "" {
			continue
		}

		apps = append(apps, model.App{
			Name:        desktop.Name,
			Command:     desktop.Exec,
			Description: desktop.Comment,
			Keywords:    desktop.Keywords,
		})
	}

	return apps, nil
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

	return desktopFiles, nil
}

func (lf *linuxFinder) GetAllAppsFolders() []string {
	folders := []string{}

	foldersFromXdg := lf.GetAppFoldersFromXdg()
	folders = append(folders, foldersFromXdg...)

	foldersDefault := lf.GetDefaultAppFolders()
	folders = append(folders, foldersDefault...)

	folders = lf.RemoveDuplicateAppFolders(folders)

	return folders
}

func (lf *linuxFinder) GetAppFoldersFromXdg() []string {
	xdg := os.Getenv("XDG_DATA_DIRS")

	folders := strings.Split(xdg, ":")

	for i := range folders {
		folders[i] = folders[i] + "/applications/"
	}

	return folders
}

func (lf *linuxFinder) GetDefaultAppFolders() []string {
	return []string{
		"/usr/share/applications/",
		"/usr/local/share/applications/",
		"/var/lib/flatpak/exports/share/applications/",
		"~/.local/share/flatpak/exports/share/application/",
	}
}

func (lf *linuxFinder) RemoveDuplicateAppFolders(folders []string) []string {
	mapFolders := map[string]bool{}

	responseFolder := []string{}
	for _, folder := range folders {
		if _, ok := mapFolders[folder]; ok {
			continue
		}

		mapFolders[folder] = true
		responseFolder = append(responseFolder, folder)
	}

	return responseFolder
}
