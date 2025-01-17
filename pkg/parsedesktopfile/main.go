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

func (p *parseDesktopFile) ParseFromString(desktop string) model.Desktop {
	return p.parse(desktop)
}

func (p *parseDesktopFile) ParseFromFile(desktopFile string) (model.Desktop, error) {
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
		line = strings.Trim(line, " ")
		line = strings.Trim(line, "\t")
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "[") {
			continue
		}

		lineArr := strings.Split(line, "=")

		mapLines[lineArr[0]] = lineArr[1]
	}

	if exec, ok := mapLines["Exec"]; ok {
		response.Exec = exec
	}
	if name, ok := mapLines["Name"]; ok {
		response.Name = name
	}
	if typeDesk, ok := mapLines["Type"]; ok {
		response.Type = typeDesk
	}
	if comment, ok := mapLines["Comment"]; ok {
		response.Comment = comment
	}
	if keywords, ok := mapLines["Keywords"]; ok {
		response.Keywords = keywords
	}
	if terminal, ok := mapLines["Terminal"]; ok {
		response.Terminal = terminal == "true"
	}

	return response
}
