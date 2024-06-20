package config

type ConfigWorker struct {
}

func (cw *ConfigWorker) GetFromFile() []Config {

	response := []Config{}

	response = append(response, Config{Title: "Mozilla Firefox", Description: "web browser", Command: "firefox"})
	response = append(response, Config{Title: "DBGate", Description: "Database IDE", Command: "flatpak run org.dbgate.DbGate"})
	response = append(response, Config{Title: "Telegram", Description: "Telegram Desktop", Command: "flatpak run org.telegram.desktop"})
	response = append(response, Config{Title: "Nemo", Description: "File manager", Command: "nemo"})
	response = append(response, Config{Title: "Project: FastLauncher", Description: "Project: FastLauncher", Command: "alacritty --working-directory ~/work/opensource/fast-launcher"})
	response = append(response, Config{Title: "Obsidian", Description: "Obsidian", Command: "flatpak run md.obsidian.Obsidian"})
	response = append(response, Config{Title: "Kate", Description: "text editor", Command: "kate"})
	response = append(response, Config{Title: "Krita", Description: "Digital painting", Command: "flatpak run org.kde.krita"})

	return response
}
