package main

import (
	"flag"

	"github.com/probeldev/fastlauncher/log"
	sourceapps "github.com/probeldev/fastlauncher/sourceApps"
	"github.com/probeldev/fastlauncher/ui"
)

func main() {

	cfgPath := flag.String("config", "", "Path to config file")
	flag.Parse()

	if cfgPath != nil && *cfgPath != "" {
		ca := sourceapps.ConfigSourceApps{}
		apps := ca.GetFromFile(*cfgPath)
		ui.StartUi(apps)
		return
	}

	oa := sourceapps.OsSourceApp{}
	apps, err := oa.GetAll()
	if err != nil {
		// TODO
		log.Println(err)
	}

	ui.StartUi(apps)

}
