package parsedesktopfile

import (
	"os"
	"strings"

	"github.com/probeldev/fastlauncher/pkg/parsedesktopfile/model"
)

type parseDesktopFile struct{}

func GetParseDesktopFile() parseDesktopFile {
	p := parseDesktopFile{}

	return p
}

func (p *parseDesktopFile) GetFromString(desktop string) model.Desktop {
	return p.parse(desktop)
}

func (p *parseDesktopFile) GetFromFile(desktopFile string) (model.Desktop, error) {
	response := model.Desktop{}
	body, err := os.ReadFile(desktopFile)

	if err != nil {
		return response, err
	}

	return p.parse(string(body)), nil
}

func (p *parseDesktopFile) parse(body string) model.Desktop {
	response := model.Desktop{}

	bodyLines := strings.Split(body, "\n")

	mapLines := map[string]string{}
	for _, line := range bodyLines {
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "[") {
			continue
		}

		lineArr := strings.Split(line, "=")

		mapLines[lineArr[0]] = lineArr[1]
	}

	if exec, ok := mapLines["exec"]; ok {
		response.Exec = exec
	}
	if name, ok := mapLines["name"]; ok {
		response.Name = name
	}
	if typeDesk, ok := mapLines["type"]; ok {
		response.Type = typeDesk
	}
	if comment, ok := mapLines["comment"]; ok {
		response.Comment = comment
	}
	if keywords, ok := mapLines["keywords"]; ok {
		response.Keywords = keywords
	}
	if terminal, ok := mapLines["terminal"]; ok {
		response.Terminal = terminal == "true"
	}

	return response
}
