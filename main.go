package main

import (
	"fast-launcher/config"
	"fast-launcher/ui"
	"flag"
)

func main() {

	cfgPath := flag.String("config", "", "Path to config file")
	flag.Parse()

	cw := config.ConfigWorker{}
	configCommands := cw.GetFromFile(*cfgPath)

	ui.StartUi(configCommands)
}
