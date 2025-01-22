package parsedesktopfile

import (
	"errors"
	"os"
	"strings"

	"github.com/probeldev/fastlauncher/pkg/parsedesktopfile/model"
)

type parseDesktopFile struct{}

func GetParseDesktopFile() parseDesktopFile {
	p := parseDesktopFile{}

	return p
}

func (p *parseDesktopFile) ParseFromString(desktop string) (
	model.Desktop,
	error,
) {
	return p.parse(desktop)
}

func (p *parseDesktopFile) ParseFromFile(desktopFile string) (
	model.Desktop,
	error,
) {
	response := model.Desktop{}
	body, err := os.ReadFile(desktopFile)

	if err != nil {
		return response, err
	}

	desktop, err := p.parse(string(body))
	if err != nil {
		return response,
			errors.New("Error parse, file: " + desktopFile + " error: " + err.Error())
	}

	return desktop, err
}

func (p *parseDesktopFile) parse(body string) (
	model.Desktop,
	error,
) {
	response := model.Desktop{}

	mapLines, err := p.GetDesktopEntry(body)
	if err != nil {
		return response, err
	}

	if exec, ok := mapLines["Exec"]; ok {
		response.Exec = strings.Trim(exec, " ")
	}
	if name, ok := mapLines["Name"]; ok {
		response.Name = strings.Trim(name, " ")
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

	return response, nil
}

func (p *parseDesktopFile) GetDesktopEntry(body string) (map[string]string, error) {

	bodyLines := strings.Split(body, "\n")

	isSetDesktopEntry := true
	responseMap := map[string]string{}
	for _, line := range bodyLines {
		line = strings.Trim(line, " ")
		line = strings.Trim(line, "\t")
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "#") {
			continue
		}

		lineArr := strings.Split(line, "=")

		if strings.HasPrefix(line, "[Desktop Entry]") {
			isSetDesktopEntry = true
			continue
		} else if strings.HasPrefix(line, "[") {
			if isSetDesktopEntry {
				return responseMap, nil
			}
			continue
		}

		if isSetDesktopEntry {
			if len(lineArr) < 2 {
				return responseMap, errors.New("Parse error, line: " + line)
			} else if len(lineArr) == 2 {
				responseMap[lineArr[0]] = lineArr[1]
			} else {
				value := lineArr[1]

				for i := 2; i < len(lineArr); i++ {
					value += "=" + lineArr[i]
				}
				responseMap[lineArr[0]] = value
			}
		}
	}

	if !isSetDesktopEntry {
		return responseMap, errors.New("Desktop Entry not found")
	}

	return responseMap, nil
}
