package mode

import (
	"log"

	flog "github.com/probeldev/fastlauncher/log"
	"github.com/probeldev/fastlauncher/model"
)

type ConfigMode struct{}

func (cw *ConfigMode) GetFromFile(cfgPath string) []model.App {
	fn := "ConfigMode:GetFromFile"

	if cfgPath == "" {
		flog.Println(fn, "cfg path not found")
		return nil
	}

	response, warnings := LoadConfigFile(cfgPath)

	// Warnings are not fatal, so the standard logger is used: the project log
	// package sleeps 15 seconds after every Println and must not delay the UI.
	for _, warning := range warnings {
		log.Println(fn, warning)
	}

	return response
}
