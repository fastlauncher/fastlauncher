package findertest_test

import (
	"testing"

	"github.com/probeldev/fastlauncher/pkg/finderallapps/finder"
)

func TestGetFromFolderNotFoundFolder(t *testing.T) {
	folder := "/not/found/folder"

	linuxFinder := finder.GetLinuxFinder()
	apps, err := linuxFinder.GetFromFolder(folder)

	if err != nil {
		t.Error(err.Error())
	}

	if len(apps) != 0 {
		t.Error("len(apps)!=0")
	}
}
