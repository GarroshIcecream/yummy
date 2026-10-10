package utils

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"charm.land/bubbles/v2/list"
	"github.com/GarroshIcecream/yummy/internal/models/common"
)

type RecipeMetadata struct {
	Author       string
	CookTime     time.Duration
	PrepTime     time.Duration
	TotalTime    time.Duration
	Quantity     string
	URL          string
	Favourite    bool
	Rating       int8
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Categories   []string
	Instructions []string
	Ingredients  []Ingredient
}

type RecipeRaw struct {
	RecipeID          uint
	RecipeName        string
	RecipeDescription string
	IsFavourite       bool
	Metadata          RecipeMetadata
}

var _ list.Item = &RecipeRaw{}

func (i RecipeRaw) Title() string {
	if i.IsFavourite {
		return "⭐ " + i.RecipeName
	}
	return i.RecipeName
}

func (i RecipeRaw) Description() string {
	if strings.TrimSpace(i.RecipeDescription) == "" {
		return i.Metadata.Author
	}

	return fmt.Sprintf("%s - %s", i.Metadata.Author, i.RecipeDescription)
}

func (i RecipeRaw) FilterValue() string {
	title := i.Title()
	ingredients := []string{}
	for _, ing := range i.Metadata.Ingredients {
		ingredients = append(ingredients, ing.Name)
	}
	filterData := map[common.FilterField]any{
		common.TitleField:       title,
		common.IngredientsField: ingredients,
		common.DescriptionField: i.RecipeDescription,
		common.AuthorField:      i.Metadata.Author,
		common.CategoryField:    i.Metadata.Categories,
		common.FavouriteField:   i.IsFavourite,
		common.URLField:         i.Metadata.URL,
	}

	jsonBytes, err := json.Marshal(filterData)
	if err != nil {
		slog.Error("Failed to marshal filter data", "error", err)
		return fmt.Sprintf("%s %s %s", i.RecipeName, i.RecipeDescription, i.Metadata.Author)
	}

	return string(jsonBytes)
}

// parseMetadataTable extracts metadata from the markdown table
func (r *RecipeRaw) parseMetadataTable(text string) {
	// Parse author
	authorMatch := regexp.MustCompile(`👨‍🍳 Recipe By\s*\|\s*(.+?)\s*\|`).FindStringSubmatch(text)
	if len(authorMatch) > 1 {
		r.Metadata.Author = strings.TrimSpace(authorMatch[1])
	}

	// Parse servings
	servingsMatch := regexp.MustCompile(`🍽️ Servings\s*\|\s*(.+?)\s*\|`).FindStringSubmatch(text)
	if len(servingsMatch) > 1 {
		r.Metadata.Quantity = strings.TrimSpace(servingsMatch[1])
	}

	// Parse times
	totalTimeMatch := regexp.MustCompile(`⏱️ Total Time\s*\|\s*(.+?)\s*\|`).FindStringSubmatch(text)
	if len(totalTimeMatch) > 1 {
		r.Metadata.TotalTime = ParseDurationFromString(strings.TrimSpace(totalTimeMatch[1]))
	}

	prepTimeMatch := regexp.MustCompile(`🔪 Prep Time\s*\|\s*(.+?)\s*\|`).FindStringSubmatch(text)
	if len(prepTimeMatch) > 1 {
		r.Metadata.PrepTime = ParseDurationFromString(strings.TrimSpace(prepTimeMatch[1]))
	}

	cookTimeMatch := regexp.MustCompile(`🔥 Cook Time\s*\|\s*(.+?)\s*\|`).FindStringSubmatch(text)
	if len(cookTimeMatch) > 1 {
		r.Metadata.CookTime = ParseDurationFromString(strings.TrimSpace(cookTimeMatch[1]))
	}
}

// formatDurationHuman formats a time.Duration into a human-friendly string
// like "45 min", "1 hr", or "1 hr 30 min". Returns "" for zero durations.
func formatDurationHuman(d time.Duration) string {
	if d <= 0 {
		return ""
	}
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	switch {
	case hours > 0 && minutes > 0:
		return fmt.Sprintf("%d hr %d min", hours, minutes)
	case hours > 0:
		return fmt.Sprintf("%d hr", hours)
	default:
		return fmt.Sprintf("%d min", minutes)
	}
}

// formatRating formats a rating (0-5) into a star string like "★★★★☆".
// Returns "" for unrated (0).
func formatRating(rating int8) string {
	if rating <= 0 {
		return ""
	}
	if rating > 5 {
		rating = 5
	}
	filled := strings.Repeat("★", int(rating))
	empty := strings.Repeat("☆", 5-int(rating))
	return filled + empty
}

// FormatRecipeMarkdown formats the recipe as markdown for the detail view,
// export and the assistant's recipe tool.
//
// It deliberately avoids emoji in structural text: multi-codepoint emoji (ZWJ
// sequences like 👨‍🍳, variation-selector forms like ⏱️) have a cell width
// terminals disagree on, which shifts the rest of the line and tears overlays
// drawn on top of it.
func (r *RecipeRaw) FormatRecipeMarkdown() string {
	var s strings.Builder

	// Title
	fmt.Fprintf(&s, "# %s\n\n", r.RecipeName)

	// At-a-glance stats line
	var stats []string
	if n := len(r.Metadata.Ingredients); n == 1 {
		stats = append(stats, "1 ingredient")
	} else if n > 1 {
		stats = append(stats, fmt.Sprintf("%d ingredients", n))
	}
	if n := len(r.Metadata.Instructions); n == 1 {
		stats = append(stats, "1 step")
	} else if n > 1 {
		stats = append(stats, fmt.Sprintf("%d steps", n))
	}
	if t := formatDurationHuman(r.Metadata.TotalTime); t != "" {
		stats = append(stats, t)
	}
	if ratingStr := formatRating(r.Metadata.Rating); ratingStr != "" {
		stats = append(stats, ratingStr)
	}
	if len(stats) > 0 {
		fmt.Fprintf(&s, "*%s*\n\n", strings.Join(stats, " | "))
	}

	// Description
	if r.RecipeDescription != "" {
		fmt.Fprintf(&s, "> %s\n\n", r.RecipeDescription)
	}

	// Metadata as a compact list of facts
	var facts []string
	addFact := func(label, value string) {
		if value != "" {
			facts = append(facts, fmt.Sprintf("- **%s** %s", label, value))
		}
	}
	addFact("By", r.Metadata.Author)
	addFact("Serves", r.Metadata.Quantity)
	addFact("Prep", formatDurationHuman(r.Metadata.PrepTime))
	addFact("Cook", formatDurationHuman(r.Metadata.CookTime))
	addFact("Total", formatDurationHuman(r.Metadata.TotalTime))
	addFact("Rating", formatRating(r.Metadata.Rating))
	if len(facts) > 0 {
		s.WriteString(strings.Join(facts, "\n") + "\n\n")
	}

	// Ingredients
	s.WriteString("## Ingredients\n\n")
	currentGroup := ""
	for _, ing := range r.Metadata.Ingredients {
		if ing.Group != currentGroup {
			currentGroup = ing.Group
			if group := strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(currentGroup), ":")); group != "" {
				fmt.Fprintf(&s, "\n### %s\n\n", group)
			}
		}

		var ingredient strings.Builder
		ingredient.WriteString("- ")

		if ing.Amount != "" && ing.Unit != "" {
			fmt.Fprintf(&ingredient, "**%s %s** ", ing.Amount, ing.Unit)
		} else if ing.Amount != "" {
			fmt.Fprintf(&ingredient, "**%s** ", ing.Amount)
		}

		ingredient.WriteString(ing.Name)

		if details := CleanIngredientDetails(ing.Details); details != "" {
			fmt.Fprintf(&ingredient, " *(%s)*", details)
		}
		s.WriteString(ingredient.String() + "\n")
	}
	s.WriteString("\n")

	// Instructions (with ingredient names highlighted)
	s.WriteString("## Instructions\n\n")
	for i, inst := range r.Metadata.Instructions {
		highlighted := HighlightIngredientsInMarkdown(inst, r.Metadata.Ingredients)
		fmt.Fprintf(&s, "**%d.** %s\n\n", i+1, highlighted)
	}

	// Categories
	if len(r.Metadata.Categories) > 0 {
		s.WriteString("## Categories\n\n")
		for _, cat := range r.Metadata.Categories {
			fmt.Fprintf(&s, "`%s` ", cat)
		}
		s.WriteString("\n\n")
	}

	// Source and timestamps
	var footer []string
	if r.Metadata.URL != "" {
		footer = append(footer, "Source: "+r.Metadata.URL)
	}
	if !r.Metadata.CreatedAt.IsZero() {
		added := "Added " + r.Metadata.CreatedAt.Format("Jan 2, 2006")
		if !r.Metadata.UpdatedAt.IsZero() && r.Metadata.UpdatedAt.Sub(r.Metadata.CreatedAt) > time.Second {
			added += " | updated " + r.Metadata.UpdatedAt.Format("Jan 2, 2006")
		}
		footer = append(footer, added)
	}
	if len(footer) > 0 {
		s.WriteString("---\n\n")
		s.WriteString(strings.Join(footer, "\n\n") + "\n")
	}

	return NormalizeSymbols(s.String())
}

// parseMarkdownRecipe parses a markdown recipe file
func ParseMarkdownRecipe(filePath string, customName string) (*RecipeRaw, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}

	text := string(content)
	recipeData := &RecipeRaw{
		Metadata: RecipeMetadata{
			Ingredients:  []Ingredient{},
			Instructions: []string{},
			Categories:   []string{},
		},
	}

	// Parse recipe name
	if customName != "" {
		recipeData.RecipeName = customName
	} else {
		nameMatch := regexp.MustCompile(`(?m)^# 🍳 (.+)$`).FindStringSubmatch(text)
		if len(nameMatch) > 1 {
			recipeData.RecipeName = strings.TrimSpace(nameMatch[1])
		} else {
			baseName := filepath.Base(filePath)
			recipeData.RecipeName = strings.TrimSuffix(baseName, filepath.Ext(baseName))
		}
	}

	// Parse description
	descMatch := regexp.MustCompile(`💭 \*About this recipe:\*\n> (.+?)\n`).FindStringSubmatch(text)
	if len(descMatch) > 1 {
		recipeData.RecipeDescription = strings.TrimSpace(descMatch[1])
	}

	// Parse metadata table
	recipeData.parseMetadataTable(text)

	// Parse ingredients
	ingredients, err := ParseIngredientsFromMarkdown(text)
	if err != nil {
		slog.Error("Failed to parse ingredients", "error", err)
	}
	recipeData.Metadata.Ingredients = ingredients

	// Parse instructions
	instructions, err := ParseInstructionsFromMarkdown(text)
	if err != nil {
		slog.Error("Failed to parse instructions", "error", err)
	}
	recipeData.Metadata.Instructions = instructions

	// Parse categories
	categories, err := ParseCategoriesFromMarkdown(text)
	if err != nil {
		slog.Error("Failed to parse categories", "error", err)
	}
	recipeData.Metadata.Categories = categories

	// Parse source URL
	URL, err := ParseSourceURLFromMarkdown(text)
	if err != nil {
		slog.Error("Failed to parse source URL", "error", err)
	}
	recipeData.Metadata.URL = URL

	return recipeData, nil
}

// parseJSONRecipe parses a JSON recipe file
func ParseJSONRecipe(filePath string, customName string) (*RecipeRaw, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}

	var jsonRecipe struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Author      string `json:"author"`
		CookTime    string `json:"cook_time"`
		PrepTime    string `json:"prep_time"`
		TotalTime   string `json:"total_time"`
		Quantity    string `json:"quantity"`
		URL         string `json:"url"`
		Ingredients []struct {
			Amount  string `json:"amount"`
			Unit    string `json:"unit"`
			Name    string `json:"name"`
			Details string `json:"details"`
			Group   string `json:"group"`
		} `json:"ingredients"`
		Instructions []string `json:"instructions"`
		Categories   []string `json:"categories"`
	}

	if err := json.Unmarshal(content, &jsonRecipe); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %v", err)
	}

	recipeData := &RecipeRaw{
		RecipeName:        jsonRecipe.Name,
		RecipeDescription: jsonRecipe.Description,
		Metadata: RecipeMetadata{
			Quantity:     jsonRecipe.Quantity,
			URL:          jsonRecipe.URL,
			Ingredients:  []Ingredient{},
			Instructions: jsonRecipe.Instructions,
			Categories:   jsonRecipe.Categories,
		},
	}

	// Override name if custom name provided
	if customName != "" {
		recipeData.RecipeName = customName
	}

	// Parse time durations
	recipeData.Metadata.CookTime = ParseDurationFromString(jsonRecipe.CookTime)
	recipeData.Metadata.PrepTime = ParseDurationFromString(jsonRecipe.PrepTime)
	recipeData.Metadata.TotalTime = ParseDurationFromString(jsonRecipe.TotalTime)

	// Convert ingredients
	for _, ing := range jsonRecipe.Ingredients {
		recipeData.Metadata.Ingredients = append(recipeData.Metadata.Ingredients, Ingredient{
			Amount:  ing.Amount,
			Unit:    ing.Unit,
			Name:    ing.Name,
			Details: ing.Details,
			Group:   ing.Group,
		})
	}

	return recipeData, nil
}

// symbolReplacer swaps letterlike symbols whose cell width terminals disagree
// on (East Asian "ambiguous" width — e.g. ℃ is drawn two cells wide by many
// terminals but measured as one) for plain equivalents, so they don't eat the
// following character or shift the rest of the line.
var symbolReplacer = strings.NewReplacer(
	"℃", "°C",
	"℉", "°F",
	"№", "No.",
	"™", "(TM)",
)

// NormalizeSymbols replaces width-ambiguous symbols with plain equivalents.
func NormalizeSymbols(s string) string {
	return symbolReplacer.Replace(s)
}

// FormatDurationHuman formats a duration as "1 hr 15 min" ("" for zero).
func FormatDurationHuman(d time.Duration) string {
	return formatDurationHuman(d)
}
