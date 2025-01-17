package sourceapps

import (
	"encoding/json"
	"os"

	"github.com/probeldev/fastlauncher/log"
	"github.com/probeldev/fastlauncher/model"
)

type ConfigSourceApps struct{}

func (cw *ConfigSourceApps) GetFromFile(cfgPath string) []model.App {

	response := []model.App{}

	if cfgPath == "" {
		log.Println("cfg path not found")
		return response
	}

	file, err := os.ReadFile(cfgPath)
	if err != nil {
		log.Println(err)
		return response
	}

	if err := json.Unmarshal(file, &response); err != nil {
		log.Println(err)
		return response
	}

	return response
}
