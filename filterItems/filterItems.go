// Package filteritems...
package filteritems

import (
	"strings"

	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/probeldev/fastlauncher/model"
)

func FilterItems(items []model.App, query string) []model.App {
	if query == "" {
		return items
	}

	query = strings.ToLower(query)
	var filtered []model.App

	for _, it := range items {
		title := strings.ToLower(it.Title)
		if fuzzy.Match(query, title) {
			filtered = append(filtered, it)
			continue
		}

		description := strings.ToLower(it.Description)
		if fuzzy.Match(query, description) {
			filtered = append(filtered, it)
			continue
		}
	}

	return filtered
}
