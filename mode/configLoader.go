package mode

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/probeldev/fastlauncher/model"
)

const (
	// maxConfigDepth limits how deep nested config files can be expanded in one branch.
	maxConfigDepth = 16
	// maxConfigExpansions limits how many config files can be expanded in total per load.
	maxConfigExpansions = 1000
)

// errExpansionBudget is returned when the shared expansion budget is exhausted.
// It is reported once via warnings; subsequent hits are silent.
var errExpansionBudget = errors.New("config expansion budget exhausted")

// configLoader expands config references and shares one budget between all branches.
type configLoader struct {
	stack        map[string]bool
	seenWarnings map[string]struct{}
	expansions   int
	budgetWarned bool
}

// warningsFor returns the warning text of err unless the same text was already
// reported. Diamond-shaped graphs may hit the same limit from many branches.
func (l *configLoader) warningsFor(err error) []string {
	// The budget error is reported once when it first happens, don't repeat it.
	if errors.Is(err, errExpansionBudget) {
		return nil
	}

	text := err.Error()

	if _, seen := l.seenWarnings[text]; seen {
		return nil
	}

	l.seenWarnings[text] = struct{}{}

	return []string{text}
}

// LoadConfigFile loads apps from a json config file and recursively expands
// every "config" reference and inline "items" list. Paths in "config" are
// resolved relative to the file that declares them ("~" is expanded to the
// home directory). Broken references, cycles, too deep nesting and exceeding
// the expansion budget are skipped and reported in the returned warnings; the
// rest of the config is still loaded.
//
// One rule decides what stays: an item that ends up with no children and no
// non-empty command is dropped. A broken reference keeps the item when it has
// inline items (as a group with those items) or a non-empty command (as a leaf).
func LoadConfigFile(cfgPath string) (apps []model.App, warnings []string) {
	path, err := resolveConfigPath(cfgPath, "")
	if err != nil {
		return nil, []string{err.Error()}
	}

	loader := &configLoader{
		stack:        map[string]bool{},
		seenWarnings: map[string]struct{}{},
	}

	loaded, loadWarnings, err := loader.loadAppsFromPath(path, 0)
	if err != nil {
		return nil, append(loadWarnings, loader.warningsFor(err)...)
	}

	return loaded, loadWarnings
}

// loadAppsFromPath reads one config file and expands its apps. The stack holds
// absolute paths of the files in the current nesting branch to detect cycles,
// while the budget caps the total number of expanded files to avoid the
// exponential blowup of diamond-shaped config graphs.
func (l *configLoader) loadAppsFromPath(
	path string,
	depth int,
) (
	[]model.App,
	[]string,
	error,
) {
	if depth >= maxConfigDepth {
		return nil, nil, fmt.Errorf(
			"config nesting depth limit (%d) exceeded: %s",
			maxConfigDepth,
			path,
		)
	}

	if l.stack[path] {
		return nil, nil, fmt.Errorf("config cycle detected: %s", path)
	}

	if l.expansions >= maxConfigExpansions {
		if l.budgetWarned {
			return nil, nil, errExpansionBudget
		}
		l.budgetWarned = true

		return nil, []string{fmt.Sprintf(
			"too many config expansions (limit %d), skipped %s",
			maxConfigExpansions,
			path,
		)}, errExpansionBudget
	}

	l.expansions++

	// Config paths are intentionally user-provided (CLI flag and config references).
	file, err := os.ReadFile(path) //nolint:gosec // user-provided config paths are the feature
	if err != nil {
		return nil, nil, fmt.Errorf("read config %s: %w", path, err)
	}

	raw, err := model.NewAppListFromJSON(file)
	if err != nil {
		return nil, nil, fmt.Errorf("parse config %s: %w", path, err)
	}

	l.stack[path] = true
	defer delete(l.stack, path)

	apps, warnings := l.expandApps(raw, filepath.Dir(path), depth)

	return apps, warnings, nil
}

// expandApps expands inline items and config references of every app.
func (l *configLoader) expandApps(
	apps []model.App,
	dir string,
	depth int,
) (
	appsExpanded []model.App,
	appWarnings []string,
) {
	expanded := make([]model.App, 0, len(apps))
	warnings := []string{}

	for i := range apps {
		item, itemWarnings := l.expandApp(&apps[i], dir, depth)
		warnings = append(warnings, itemWarnings...)

		if item == nil {
			continue
		}

		expanded = append(expanded, *item)
	}

	return expanded, warnings
}

// expandApp expands one app and applies the single rule: an item without
// children and without a non-empty command is dropped (returns nil).
func (l *configLoader) expandApp(
	item *model.App,
	dir string,
	depth int,
) (
	expandedApp *model.App,
	appWarnings []string,
) {
	inlineItems, warnings := l.expandApps(item.Items, dir, depth)

	cfgPath := item.Config

	expanded := *item
	expanded.Config = ""
	expanded.Items = inlineItems

	if len(expanded.Items) == 0 {
		expanded.Items = nil
	}

	if cfgPath == "" {
		return finishApp(&expanded, warnings)
	}

	resolved, err := resolveConfigPath(cfgPath, dir)
	if err != nil {
		warnings = append(warnings, l.warningsFor(err)...)

		return finishApp(&expanded, warnings)
	}

	loaded, loadWarnings, err := l.loadAppsFromPath(resolved, depth+1)
	warnings = append(warnings, loadWarnings...)

	if err != nil {
		warnings = append(warnings, l.warningsFor(err)...)

		return finishApp(&expanded, warnings)
	}

	expanded.Items = append(expanded.Items, loaded...)

	return finishApp(&expanded, warnings)
}

// finishApp applies the single rule for keeping items: an item with children
// is a group, an item with a non-empty command is a leaf, anything else is
// dropped. It also normalizes an empty children slice to nil.
func finishApp(
	item *model.App,
	warnings []string,
) (
	keptItem *model.App,
	keptWarnings []string,
) {
	if len(item.Items) == 0 {
		item.Items = nil
	}

	if len(item.Items) == 0 && strings.TrimSpace(item.Command) == "" {
		return nil, warnings
	}

	return item, warnings
}

// resolveConfigPath expands "~" and resolves relative paths against baseDir.
func resolveConfigPath(cfgPath, baseDir string) (string, error) {
	path := cfgPath

	if path == "~" || strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolve home directory for %q: %w", cfgPath, err)
		}

		path = filepath.Join(home, strings.TrimPrefix(path, "~"))
	}

	if !filepath.IsAbs(path) {
		path = filepath.Join(baseDir, path)
	}

	abs, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve config path %q: %w", cfgPath, err)
	}

	return abs, nil
}
