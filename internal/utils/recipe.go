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

	"github.com/GarroshIcecream/yummy/internal/models/common"
	"github.com/charmbracelet/bubbles/list"
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

// FormatRecipeContent formats the recipe content into a markdown string
func (r *RecipeRaw) FormatRecipeMarkdown() string {
	var s strings.Builder

	// Title
	s.WriteString(fmt.Sprintf("# 🍳 %s\n\n", r.RecipeName))

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
	if len(stats) > 0 {
		s.WriteString(fmt.Sprintf("*%s*\n\n", strings.Join(stats, " • ")))
	}

	// Description
	if r.RecipeDescription != "" {
		s.WriteString("💭 *About this recipe:*\n")
		s.WriteString(fmt.Sprintf("> %s\n\n", r.RecipeDescription))
	}

	// Metadata as simple key-value pairs (no table)
	metaRows := []struct{ label, value string }{}
	if r.Metadata.Author != "" {
		metaRows = append(metaRows, struct{ label, value string }{"👨‍🍳 Recipe By", r.Metadata.Author})
	}
	if r.Metadata.Quantity != "" {
		metaRows = append(metaRows, struct{ label, value string }{"🍽️ Servings", r.Metadata.Quantity})
	}
	if t := formatDurationHuman(r.Metadata.TotalTime); t != "" {
		metaRows = append(metaRows, struct{ label, value string }{"⏱️ Total Time", t})
	}
	if t := formatDurationHuman(r.Metadata.PrepTime); t != "" {
		metaRows = append(metaRows, struct{ label, value string }{"🔪 Prep Time", t})
	}
	if t := formatDurationHuman(r.Metadata.CookTime); t != "" {
		metaRows = append(metaRows, struct{ label, value string }{"🔥 Cook Time", t})
	}
	if ratingStr := formatRating(r.Metadata.Rating); ratingStr != "" {
		metaRows = append(metaRows, struct{ label, value string }{"⭐ Rating", ratingStr})
	}

	if len(metaRows) > 0 {
		s.WriteString("---\n\n")
		for _, row := range metaRows {
			s.WriteString(fmt.Sprintf("  %s  **%s**\n\n", row.label, row.value))
		}
		s.WriteString("---\n\n")
	}

	// Ingredients
	s.WriteString("### 🥘 Ingredients\n\n")
	for _, ing := range r.Metadata.Ingredients {
		var ingredient strings.Builder
		ingredient.WriteString("• ")

		if ing.Amount != "" && ing.Unit != "" {
			ingredient.WriteString(fmt.Sprintf("**%s %s** ", ing.Amount, ing.Unit))
		} else if ing.Amount != "" {
			ingredient.WriteString(fmt.Sprintf("**%s** ", ing.Amount))
		}

		ingredient.WriteString(fmt.Sprintf("*%s*", ing.Name))

		if ing.Details != "" {
			ingredient.WriteString(fmt.Sprintf(" (%s)", ing.Details))
		}
		s.WriteString(ingredient.String() + "\n\n")
	}

	// Instructions (with ingredient names highlighted)
	s.WriteString("### 👩‍🍳 Instructions\n\n")
	for i, inst := range r.Metadata.Instructions {
		highlighted := HighlightIngredientsInMarkdown(inst, r.Metadata.Ingredients)
		s.WriteString(fmt.Sprintf("**%d.** %s\n\n", i+1, highlighted))
	}

	// Categories
	if len(r.Metadata.Categories) > 0 {
		s.WriteString("### 🏷️ Categories\n\n")
		for _, cat := range r.Metadata.Categories {
			s.WriteString(fmt.Sprintf("`%s` ", cat))
		}
		s.WriteString("\n\n")
	}

	// Source
	if r.Metadata.URL != "" {
		s.WriteString("🔗 " + r.Metadata.URL + "\n\n")
	}

	// Timestamps
	if !r.Metadata.CreatedAt.IsZero() {
		s.WriteString(fmt.Sprintf("📅 Added on %s", r.Metadata.CreatedAt.Format("Jan 2, 2006")))
		if !r.Metadata.UpdatedAt.IsZero() && r.Metadata.UpdatedAt.Sub(r.Metadata.CreatedAt) > time.Second {
			s.WriteString(fmt.Sprintf("  ·  🔄 Updated %s", r.Metadata.UpdatedAt.Format("Jan 2, 2006")))
		}
		s.WriteString("\n")
	}

	return s.String()
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
		})
	}

	return recipeData, nil
}
