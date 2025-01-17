package main

import (
	"fastlauncher/config"
	"fastlauncher/ui"
	"flag"
)

func main() {

	cfgPath := flag.String("config", "", "Path to config file")
	flag.Parse()

	cw := config.ConfigWorker{}
	configCommands := cw.GetFromFile(*cfgPath)

	ui.StartUi(configCommands)
}
