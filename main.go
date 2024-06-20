package main

import (
	"fast-launcher/config"
	"fast-launcher/ui"
)

func main() {

	cw := config.ConfigWorker{}
	configCommands := cw.GetFromFile()

	ui.StartUi(configCommands)
}
