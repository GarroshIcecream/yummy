package tools

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"strings"

	"github.com/GarroshIcecream/yummy/internal/db"
	"github.com/tmc/langchaingo/tools"
)

// CookbookOverviewTool summarises what the cookbook contains, so the assistant
// can suggest valid filters instead of guessing category names.
type CookbookOverviewTool struct {
	Cookbook *db.CookBook
}

var _ tools.Tool = &CookbookOverviewTool{}

func NewCookbookOverviewTool(cookbook *db.CookBook) *CookbookOverviewTool {
	return &CookbookOverviewTool{Cookbook: cookbook}
}

func (t *CookbookOverviewTool) Name() string { return "getCookbookOverview" }

func (t *CookbookOverviewTool) Description() string {
	return "Summarise the cookbook: number of recipes, favourites, and the categories and authors with recipe counts. " +
		"Use it to learn valid filter values for listRecipes. Input: none (send an empty string)."
}

func (t *CookbookOverviewTool) Call(ctx context.Context, input string) (string, error) {
	slog.Debug("Executing tool", "tool", t.Name(), "input", input)
	recipes, err := t.Cookbook.AllRecipes()
	if err != nil {
		return "", fmt.Errorf("failed to fetch recipes: %w", err)
	}
	if len(recipes) == 0 {
		return "The cookbook is empty.", nil
	}

	categories := map[string]int{}
	authors := map[string]int{}
	favourites := 0
	for _, r := range recipes {
		for _, c := range r.Metadata.Categories {
			categories[c]++
		}
		if r.Metadata.Author != "" {
			authors[r.Metadata.Author]++
		}
		if r.IsFavourite || r.Metadata.Favourite {
			favourites++
		}
	}

	var b strings.Builder
	fmt.Fprintf(&b, "%d recipes, %d favourite(s).\n", len(recipes), favourites)
	fmt.Fprintf(&b, "Categories: %s\n", countList(categories))
	fmt.Fprintf(&b, "Authors: %s\n", countList(authors))
	return b.String(), nil
}

func countList(m map[string]int) string {
	if len(m) == 0 {
		return "none"
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if m[keys[i]] != m[keys[j]] {
			return m[keys[i]] > m[keys[j]]
		}
		return keys[i] < keys[j]
	})
	parts := make([]string, len(keys))
	for i, k := range keys {
		parts[i] = fmt.Sprintf("%s (%d)", k, m[k])
	}
	return strings.Join(parts, ", ")
}
