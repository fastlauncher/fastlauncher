package config

import (
	"encoding/json"
	"os"

	"github.com/probeldev/fastlauncher/log"
)

type ConfigWorker struct {
}

func (cw *ConfigWorker) GetFromFile(cfgPath string) []Config {

	response := []Config{}

	if cfgPath == "" {
		log.Println("cfg path not found")
		return response
	}

	file, err := os.ReadFile(cfgPath)
	if err != nil {
		log.Println("Ошибка при чтении файла:", err)
		return response
	}

	if err := json.Unmarshal(file, &response); err != nil {
		log.Println("Ошибка при десериализации JSON:", err)
		return response
	}

	return response
}
