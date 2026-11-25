package mode_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/probeldev/fastlauncher/mode"
	"github.com/probeldev/fastlauncher/model"
)

func writeConfig(t *testing.T, dir, name, body string) string {
	t.Helper()

	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}

	return path
}

func appTitles(apps []model.App) []string {
	titles := make([]string, 0, len(apps))
	for _, app := range apps {
		titles = append(titles, app.Title)
	}

	return titles
}

func findApp(t *testing.T, apps []model.App, title string) model.App {
	t.Helper()

	for _, app := range apps {
		if app.Title == title {
			return app
		}
	}

	t.Fatalf("app %q not found in %v", title, appTitles(apps))

	return model.App{}
}

func TestLoadFlatConfig(t *testing.T) {
	dir := t.TempDir()
	cfg := writeConfig(t, dir, "cfg.json", `[
		{"title": "Safari", "description": "web browser", "command": "open -n Safari.app"},
		{"title": "Editor", "description": "text editor", "command": "open -n Editor.app"}
	]`)

	apps, warnings := mode.LoadConfigFile(cfg)

	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %v", warnings)
	}
	if len(apps) != 2 {
		t.Fatalf("expected 2 apps, got %d", len(apps))
	}
	if apps[0].Command != "open -n Safari.app" || apps[1].Title != "Editor" {
		t.Fatalf("unexpected apps: %+v", apps)
	}
	if apps[0].Items != nil || apps[0].Config != "" {
		t.Fatalf("flat config must not have nested fields: %+v", apps[0])
	}
}

func TestLoadInlineItems(t *testing.T) {
	dir := t.TempDir()
	cfg := writeConfig(t, dir, "cfg.json", `[
		{
			"title": "Dev",
			"description": "development",
			"command": "ignored",
			"items": [
				{"title": "Project A", "description": "first", "command": "code a"},
				{"title": "Project B", "description": "second", "command": "code b"}
			]
		}
	]`)

	apps, warnings := mode.LoadConfigFile(cfg)

	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %v", warnings)
	}
	if len(apps) != 1 {
		t.Fatalf("expected 1 app, got %d", len(apps))
	}

	group := apps[0]
	if len(group.Items) != 2 {
		t.Fatalf("expected 2 nested items, got %d", len(group.Items))
	}
	if group.Items[1].Command != "code b" {
		t.Fatalf("nested command lost: %+v", group.Items[1])
	}
}

func TestLoadConfigRelativeToFile(t *testing.T) {
	dir := t.TempDir()
	writeConfig(t, dir, "dev.json", `[
		{"title": "Project A", "description": "first", "command": "code a"}
	]`)
	cfg := writeConfig(t, dir, "cfg.json", `[
		{"title": "Dev", "description": "development", "config": "dev.json"},
		{"title": "Safari", "description": "web browser", "command": "open Safari.app"}
	]`)

	apps, warnings := mode.LoadConfigFile(cfg)

	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %v", warnings)
	}

	group := findApp(t, apps, "Dev")
	if len(group.Items) != 1 || group.Items[0].Title != "Project A" {
		t.Fatalf("config reference was not expanded: %+v", group)
	}
	if group.Config != "" {
		t.Fatalf("Config must be cleared after expansion: %+v", group)
	}
	findApp(t, apps, "Safari")
}

func TestLoadConfigFromHome(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	writeConfig(t, home, "prompt.json", `[
		{"title": "Prompt A", "description": "prompt", "command": "pbcopy a"}
	]`)
	cfg := writeConfig(t, t.TempDir(), "cfg.json", `[
		{"title": "LLM", "description": "prompts", "config": "~/prompt.json"}
	]`)

	apps, warnings := mode.LoadConfigFile(cfg)

	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %v", warnings)
	}

	group := findApp(t, apps, "LLM")
	if len(group.Items) != 1 || group.Items[0].Title != "Prompt A" {
		t.Fatalf("home reference was not expanded: %+v", group)
	}
}

func TestLoadConfigTwoLevels(t *testing.T) {
	dir := t.TempDir()
	writeConfig(t, dir, "leaf.json", `[
		{"title": "Leaf", "description": "leaf", "command": "run leaf"}
	]`)
	writeConfig(t, dir, "middle.json", `[
		{"title": "Middle", "description": "middle", "config": "leaf.json"}
	]`)
	cfg := writeConfig(t, dir, "cfg.json", `[
		{"title": "Top", "description": "top", "config": "middle.json"}
	]`)

	apps, warnings := mode.LoadConfigFile(cfg)

	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %v", warnings)
	}

	top := findApp(t, apps, "Top")
	middle := findApp(t, top.Items, "Middle")
	leaf := findApp(t, middle.Items, "Leaf")
	if leaf.Command != "run leaf" {
		t.Fatalf("unexpected leaf: %+v", leaf)
	}
}

func TestLoadConfigCycle(t *testing.T) {
	dir := t.TempDir()
	writeConfig(t, dir, "a.json", `[
		{"title": "A", "description": "a", "config": "b.json"}
	]`)
	writeConfig(t, dir, "b.json", `[
		{"title": "B", "description": "b", "config": "a.json"}
	]`)

	apps, warnings := mode.LoadConfigFile(filepath.Join(dir, "a.json"))

	if len(warnings) == 0 {
		t.Fatal("expected cycle warning")
	}
	if !strings.Contains(strings.Join(warnings, "\n"), "cycle detected") {
		t.Fatalf("expected cycle warning, got: %v", warnings)
	}
	// Both items are shells without children or commands, so they are skipped.
	if len(apps) != 0 {
		t.Fatalf("cyclic items must be skipped: %v", appTitles(apps))
	}
}

func TestLoadConfigMissingFile(t *testing.T) {
	dir := t.TempDir()
	cfg := writeConfig(t, dir, "cfg.json", `[
		{"title": "Before", "description": "before", "command": "run before"},
		{"title": "Broken", "description": "broken", "config": "missing.json"},
		{"title": "After", "description": "after", "command": "run after"}
	]`)

	apps, warnings := mode.LoadConfigFile(cfg)

	if len(warnings) == 0 {
		t.Fatal("expected warning about missing file")
	}
	if len(apps) != 2 {
		t.Fatalf("expected 2 apps, got %v", appTitles(apps))
	}
	if apps[0].Title != "Before" || apps[1].Title != "After" {
		t.Fatalf("broken item must be skipped: %v", appTitles(apps))
	}
}

func TestLoadConfigInsideInlineItemsRelativeToFile(t *testing.T) {
	dir := t.TempDir()
	writeConfig(t, dir, "child.json", `[
		{"title": "Child From Temp Dir", "description": "child", "command": "run child"}
	]`)
	cfg := writeConfig(t, dir, "cfg.json", `[
		{
			"title": "Group",
			"description": "group",
			"items": [
				{"title": "Nested", "description": "nested", "config": "child.json"}
			]
		}
	]`)

	apps, warnings := mode.LoadConfigFile(cfg)

	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %v", warnings)
	}

	group := findApp(t, apps, "Group")
	nested := findApp(t, group.Items, "Nested")
	child := findApp(t, nested.Items, "Child From Temp Dir")
	if child.Command != "run child" {
		t.Fatalf("unexpected child: %+v", child)
	}
}

func TestLoadConfigDepthLimit(t *testing.T) {
	dir := t.TempDir()

	// Chain longer than the depth limit: 0.json -> 1.json -> ... -> 19.json.
	// Every file has a valid leaf and, except the last one, a reference to the
	// next file. So the chain does not collapse into empty shells and the depth
	// of the resulting tree is bounded by maxConfigDepth.
	const files = 20
	for i := files - 1; i >= 0; i-- {
		body := `[{"title": "Leaf", "description": "leaf", "command": "run leaf"}]`
		if i < files-1 {
			body = `[
				{"title": "Leaf", "description": "leaf", "command": "run leaf"},
				{"title": "Level", "description": "level", "config": "` +
				fmt.Sprintf("%d.json", i+1) + `"}
			]`
		}
		writeConfig(t, dir, fmt.Sprintf("%d.json", i), body)
	}

	apps, warnings := mode.LoadConfigFile(filepath.Join(dir, "0.json"))

	depthWarnings := 0
	for _, warning := range warnings {
		if strings.Contains(warning, "depth limit") {
			depthWarnings++
		}
	}
	if depthWarnings != 1 {
		t.Fatalf("expected exactly 1 depth limit warning, got %d: %v", depthWarnings, warnings)
	}

	depth := 0
	current := apps
	for len(current) > 0 {
		depth++
		current = current[len(current)-1].Items
	}
	if depth != 16 {
		t.Fatalf("expected nesting depth 16, got %d", depth)
	}
}

func TestLoadConfigBrokenNestedJSON(t *testing.T) {
	dir := t.TempDir()
	writeConfig(t, dir, "broken.json", `[{"title": "Broken",`)
	cfg := writeConfig(t, dir, "cfg.json", `[
		{"title": "Before", "description": "before", "command": "run before"},
		{"title": "Broken Group", "description": "broken", "config": "broken.json"},
		{"title": "After", "description": "after", "command": "run after"}
	]`)

	apps, warnings := mode.LoadConfigFile(cfg)

	if len(warnings) == 0 {
		t.Fatal("expected warning about broken nested json")
	}
	if len(apps) != 2 {
		t.Fatalf("expected 2 apps, got %v", appTitles(apps))
	}
	if apps[0].Title != "Before" || apps[1].Title != "After" {
		t.Fatalf("broken item must be skipped: %v", appTitles(apps))
	}
}

func TestLoadBrokenTopLevelJSON(t *testing.T) {
	dir := t.TempDir()
	cfg := writeConfig(t, dir, "cfg.json", `[{"title": "Broken",`)

	apps, warnings := mode.LoadConfigFile(cfg)

	if len(apps) != 0 {
		t.Fatalf("expected no apps, got %v", appTitles(apps))
	}
	if len(warnings) == 0 {
		t.Fatal("expected warning about broken top level json")
	}
}

func TestLoadConfigExpansionBudget(t *testing.T) {
	dir := t.TempDir()

	// Wide diamond-shaped graph: every file references the next one twice,
	// which would blow up exponentially without a shared budget.
	const files = 30
	for i := files - 1; i >= 0; i-- {
		body := `[{"title": "Leaf", "description": "leaf", "command": "run leaf"}]`
		if i < files-1 {
			ref := fmt.Sprintf("%d.json", i+1)
			body = `[
				{"title": "Left", "description": "left", "config": "` + ref + `"},
				{"title": "Right", "description": "right", "config": "` + ref + `"},
				{"title": "Sibling", "description": "sibling", "command": "run sibling"}
			]`
		}
		writeConfig(t, dir, fmt.Sprintf("%d.json", i), body)
	}

	apps, warnings := mode.LoadConfigFile(filepath.Join(dir, "0.json"))

	if len(warnings) == 0 {
		t.Fatal("expected expansion budget warning")
	}
	budgetWarnings := 0
	for _, warning := range warnings {
		if strings.Contains(warning, "too many config expansions") {
			budgetWarnings++
		}
	}
	if budgetWarnings != 1 {
		t.Fatalf("expected exactly 1 budget warning, got %d: %v", budgetWarnings, warnings)
	}
	// Items that hit the exhausted budget and lost all children are dropped;
	// the first branch (with its leaf siblings) still loads.
	if len(apps) == 0 {
		t.Fatal("top level items must survive budget exhaustion")
	}
}

func TestLoadConfigKeepsInlineItemsOnBrokenReference(t *testing.T) {
	dir := t.TempDir()
	cfg := writeConfig(t, dir, "cfg.json", `[
		{
			"title": "Mixed",
			"description": "inline plus broken reference",
			"config": "missing.json",
			"items": [
				{"title": "Inline", "description": "inline", "command": "run inline"}
			]
		}
	]`)

	apps, warnings := mode.LoadConfigFile(cfg)

	if len(warnings) == 0 {
		t.Fatal("expected warning about missing reference")
	}

	group := findApp(t, apps, "Mixed")
	if len(group.Items) != 1 || group.Items[0].Title != "Inline" {
		t.Fatalf("inline items must survive a broken reference: %+v", group)
	}
}

func TestLoadConfigKeepsCommandOnBrokenReference(t *testing.T) {
	dir := t.TempDir()
	cfg := writeConfig(t, dir, "cfg.json", `[
		{
			"title": "Group With Command",
			"description": "broken reference, but the command still works",
			"command": "run fallback",
			"config": "missing.json"
		}
	]`)

	apps, warnings := mode.LoadConfigFile(cfg)

	if len(warnings) == 0 {
		t.Fatal("expected warning about missing reference")
	}

	app := findApp(t, apps, "Group With Command")
	if len(app.Items) != 0 {
		t.Fatalf("expected a leaf without children, got %+v", app.Items)
	}
	if app.Command != "run fallback" {
		t.Fatalf("command must be kept: %+v", app)
	}
}

func TestLoadDropsEmptyShells(t *testing.T) {
	dir := t.TempDir()
	writeConfig(t, dir, "broken.json", `[{"title": "Broken",`)
	cfg := writeConfig(t, dir, "cfg.json", `[
		{"title": "Empty inline group", "description": "all children broken", "items": [
			{"title": "Nested broken", "description": "nested", "config": "broken.json"}
		]},
		{"title": "Empty config group", "description": "all children broken", "config": "broken.json"},
		{"title": "Blank command", "description": "only spaces", "command": "   "},
		{"title": "Alive", "description": "alive", "command": "run alive"}
	]`)

	apps, warnings := mode.LoadConfigFile(cfg)

	if len(warnings) == 0 {
		t.Fatal("expected warnings about broken nested json")
	}
	if len(apps) != 1 || apps[0].Title != "Alive" {
		t.Fatalf("empty shells must be dropped, got %v", appTitles(apps))
	}
}

func TestLoadConfigDeduplicatesRepeatedWarnings(t *testing.T) {
	// The same broken reference in neighbouring items must produce one warning.
	t.Run("read error", func(t *testing.T) {
		dir := t.TempDir()
		cfg := writeConfig(t, dir, "cfg.json", `[
			{"title": "First", "description": "first", "config": "missing.json"},
			{"title": "Second", "description": "second", "config": "missing.json"}
		]`)

		assertWarningOnce(t, cfg, "read config")
	})

	// An empty HOME makes path resolution fail, which is a separate warning
	// path (resolveConfigPath) and must be deduplicated as well.
	t.Run("resolve error", func(t *testing.T) {
		t.Setenv("HOME", "")

		dir := t.TempDir()
		cfg := writeConfig(t, dir, "cfg.json", `[
			{"title": "First", "description": "first", "config": "~/missing.json"},
			{"title": "Second", "description": "second", "config": "~/missing.json"}
		]`)

		assertWarningOnce(t, cfg, "resolve home directory")
	})
}

func assertWarningOnce(t *testing.T, cfg, substr string) {
	t.Helper()

	apps, warnings := mode.LoadConfigFile(cfg)

	if len(apps) != 0 {
		t.Fatalf("expected no apps, got %v", appTitles(apps))
	}

	count := 0
	for _, warning := range warnings {
		if strings.Contains(warning, substr) {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("expected the same warning once, got %d: %v", count, warnings)
	}
}
