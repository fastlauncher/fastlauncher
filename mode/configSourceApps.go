package mode

import (
	"os"

	"github.com/probeldev/fastlauncher/log"
	"github.com/probeldev/fastlauncher/model"
)

type ConfigMode struct{}

func (cw *ConfigMode) GetFromFile(cfgPath string) []model.App {
	fn := "ConfigMode:GetFromFile"

	if cfgPath == "" {
		log.Println(fn, "cfg path not found")
		return nil
	}

	file, err := os.ReadFile(cfgPath)
	if err != nil {
		log.Println(fn, err)
		return nil
	}

	response, err := model.NewAppListFromJSON(file)
	if err != nil {
		log.Println(fn, err)
		return nil
	}

	return response
}
