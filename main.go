package main

import (
	"flag"

	"github.com/probeldev/fastlauncher/config"
	"github.com/probeldev/fastlauncher/ui"
)

func main() {

	cfgPath := flag.String("config", "", "Path to config file")
	flag.Parse()

	cw := config.ConfigWorker{}
	configCommands := cw.GetFromFile(*cfgPath)

	ui.StartUi(configCommands)
}
