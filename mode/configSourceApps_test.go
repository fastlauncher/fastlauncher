package mode_test

import (
	"testing"
	"time"

	"github.com/probeldev/fastlauncher/mode"
)

// The project log package sleeps 15 seconds after every Println. Warning
// logging must not go through it, otherwise a config with N broken references
// would delay the launcher by N*15 seconds.
func TestGetFromFileBrokenReferencesAreFast(t *testing.T) {
	dir := t.TempDir()
	cfg := writeConfig(t, dir, "cfg.json", `[
		{"title": "One", "description": "one", "command": "run one"},
		{"title": "Broken A", "description": "a", "config": "missing-a.json"},
		{"title": "Broken B", "description": "b", "config": "missing-b.json"},
		{"title": "Broken C", "description": "c", "config": "missing-c.json"},
		{"title": "Two", "description": "two", "command": "run two"}
	]`)

	start := time.Now()
	apps := (&mode.ConfigMode{}).GetFromFile(cfg)
	elapsed := time.Since(start)

	if elapsed >= time.Second {
		t.Fatalf("GetFromFile took %v, warnings must not use the sleeping log package", elapsed)
	}
	if len(apps) != 2 {
		t.Fatalf("expected 2 apps, got %d", len(apps))
	}
}

func TestGetFromFileLoadsNested(t *testing.T) {
	dir := t.TempDir()
	writeConfig(t, dir, "child.json", `[
		{"title": "Child", "description": "child", "command": "run child"}
	]`)
	cfg := writeConfig(t, dir, "cfg.json", `[
		{"title": "Group", "description": "group", "config": "child.json"}
	]`)

	apps := (&mode.ConfigMode{}).GetFromFile(cfg)

	if len(apps) != 1 || len(apps[0].Items) != 1 {
		t.Fatalf("nested config was not expanded: %+v", apps)
	}
	if apps[0].Items[0].Title != "Child" {
		t.Fatalf("unexpected child: %+v", apps[0].Items)
	}
}
