package parsedesktopfile_test

import (
	"testing"

	"github.com/probeldev/fastlauncher/pkg/parsedesktopfile"
	"github.com/probeldev/fastlauncher/pkg/parsedesktopfile/model"
)

func TestParseFromString(t *testing.T) {
	fn := "TestParseFromString"
	body := `
[Desktop Entry]
Type=Application
Exec=3foot
Icon=foot
Terminal=false
Categories=System;TerminalEmulator;
Keywords=shell;prompt;command;commandline;

Name=Foot
GenericName=Terminal
Comment=A wayland native terminal emulator
	`
	expected := model.Desktop{
		Type: "Application",
		Exec: "foot",

		Terminal: false,
		Keywords: "shell;prompt;command;commandline;",

		Name:    "Foot",
		Comment: "A wayland native terminal emulator",
	}

	parse := parsedesktopfile.GetParseDesktopFile()

	actual := parse.ParseFromString(body)

	if actual != expected {
		t.Errorf(fn, "not equal", actual, parse)
	}

}
