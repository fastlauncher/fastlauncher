package parsedesktopfile_test

import (
	"testing"

	"github.com/probeldev/fastlauncher/pkg/parsedesktopfile"
	"github.com/probeldev/fastlauncher/pkg/parsedesktopfile/model"
)

func TestParseFromString(t *testing.T) {
	body := `
[Desktop Entry]
Type=Application
Exec=foot
Icon=foot
Terminal=false
Categories=System;TerminalEmulator;
Keywords=shell;prompt;command;commandline;

# comment

Name=ImageMagick (color depth=q16)
GenericName=Terminal
Comment=A wayland native terminal emulator
	`
	expected := model.Desktop{
		Type: "Application",
		Exec: "foot",

		Terminal: false,
		Keywords: "shell;prompt;command;commandline;",

		Name:    "ImageMagick (color depth=q16)",
		Comment: "A wayland native terminal emulator",
	}

	parse := parsedesktopfile.GetParseDesktopFile()

	actual, err := parse.ParseFromString(body)

	if err != nil {
		t.Error(err.Error())
	}

	if actual != expected {
		t.Error("not equal", actual, parse)
	}

}

func TestGetDesktopEntry(t *testing.T) {
	body := `
[Desktop Entry]
Version=1.0
Terminal=false
NoDisplay=false
Icon=org.libreoffice.LibreOffice.startcenter
Type=Application

[Desktop Action Writer]
Name=Writer
Exec=/usr/bin/flatpak run --branch=stable --arch=x86_64 --command=libreoffice org.libreoffice.LibreOffice --writer

[Desktop Action Calc]
Name=Calc
Exec=/usr/bin/flatpak run --branch=stable --arch=x86_64 --command=libreoffice org.libreoffice.LibreOffice --calc

[Desktop Action Impress]
Name=Impress
Exec=/usr/bin/flatpak run --branch=stable --arch=x86_64 --command=libreoffice org.libreoffice.LibreOffice --impress

[Desktop Action Draw]
Name=Draw
Exec=/usr/bin/flatpak run --branch=stable --arch=x86_64 --command=libreoffice org.libreoffice.LibreOffice --draw
	`

	expected := map[string]string{}
	expected["Version"] = "1.0"
	expected["Terminal"] = "false"
	expected["NoDisplay"] = "false"
	expected["Icon"] = "org.libreoffice.LibreOffice.startcenter"
	expected["Type"] = "Application"

	parser := parsedesktopfile.GetParseDesktopFile()

	actual, err := parser.GetDesktopEntry(body)

	if err != nil {
		t.Error(err.Error())
		return
	}

	if len(expected) != len(actual) {
		t.Error("not equal", actual, expected)
		return
	}

	for key, value := range expected {
		valueActual, ok := actual[key]
		if !ok {
			t.Error("not equal", actual, expected)
			return
		}

		if value != valueActual {
			t.Error("not equal", actual, expected)
			return
		}
	}

}
