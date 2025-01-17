package main

import (
	"flag"

	sourceapps "github.com/probeldev/fastlauncher/sourceApps"
	"github.com/probeldev/fastlauncher/ui"
)

func main() {

	cfgPath := flag.String("config", "", "Path to config file")
	flag.Parse()

	ca := sourceapps.ConfigSourceApps{}
	configCommands := ca.GetFromFile(*cfgPath)

	ui.StartUi(configCommands)
}
