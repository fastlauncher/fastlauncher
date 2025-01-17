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
