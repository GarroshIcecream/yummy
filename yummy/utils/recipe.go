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

	"github.com/GarroshIcecream/yummy/yummy/models/common"
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

// FormatRecipeContent formats the recipe content into a markdown string
func (r *RecipeRaw) FormatRecipeMarkdown() string {
	var s strings.Builder

	s.WriteString(fmt.Sprintf("# 🍳 %s\n\n", r.RecipeName))

	if r.RecipeDescription != "" {
		s.WriteString("💭 *About this recipe:*\n")
		s.WriteString(fmt.Sprintf("> %s\n\n", r.RecipeDescription))
	}

	s.WriteString("## 👩‍🍳 Kitchen Prep\n\n")
	recipe_table := []struct {
		title  string
		value  string
		length int
	}{
		{title: "👨‍🍳 Recipe By", value: r.Metadata.Author, length: len(r.Metadata.Author)},
		{title: "🍽️ Servings", value: r.Metadata.Quantity, length: len(r.Metadata.Quantity)},
		{title: "⏱️ Total Time", value: r.Metadata.TotalTime.String(), length: len(r.Metadata.TotalTime.String())},
		{title: "🔪 Prep Time", value: r.Metadata.PrepTime.String(), length: len(r.Metadata.PrepTime.String())},
		{title: "🔥 Cook Time", value: r.Metadata.CookTime.String(), length: len(r.Metadata.CookTime.String())},
	}

	longest_string := 0
	for _, item := range recipe_table {
		if item.length > longest_string {
			longest_string = item.length
		}
	}

	first_column_length := 15
	table_first_col_name := "Metadata"
	table_second_col_name := "Details"
	s.WriteString(fmt.Sprintf("| %s | %s |\n", table_first_col_name, table_second_col_name))
	s.WriteString(fmt.Sprintf("| %s | %s |\n", strings.Repeat("-", first_column_length), strings.Repeat("-", longest_string)))

	for _, item := range recipe_table {
		r_pad_value := longest_string - item.length
		r_pad_title := first_column_length - len(item.title)

		// Ensure padding values are never negative
		if r_pad_value < 0 {
			r_pad_value = 0
		}
		if r_pad_title < 0 {
			r_pad_title = 0
		}

		s.WriteString(fmt.Sprintf("| %s | %s |\n", item.title+strings.Repeat(" ", r_pad_title), item.value+strings.Repeat(" ", r_pad_value)))
	}
	s.WriteString("\n")

	s.WriteString("## 🥘 Ingredients\n\n")
	s.WriteString("*Gather all ingredients before starting:*\n\n")

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
	s.WriteString("\n")

	// Instructions section with clean formatting
	s.WriteString("## 👩‍🍳 Cooking Instructions\n\n")
	s.WriteString("*Follow these steps in order:*\n\n")

	for i, inst := range r.Metadata.Instructions {
		s.WriteString(fmt.Sprintf("%d. %s\n\n", i+1, inst))
	}

	// Categories and Tags
	if len(r.Metadata.Categories) > 0 {
		s.WriteString("## 🏷️ Recipe Type\n\n")
		s.WriteString("*This recipe falls under:*\n\n")
		for _, cat := range r.Metadata.Categories {
			s.WriteString(fmt.Sprintf("`%s` ", cat))
		}
		s.WriteString("\n\n")
	}

	// Source Attribution
	if r.Metadata.URL != "" {
		s.WriteString("## 📖 Recipe Source\n\n")
		s.WriteString("*Want to learn more? Check out the original recipe:*\n\n")
		s.WriteString(fmt.Sprintf("🔗 [View Original Recipe](%s)\n\n", r.Metadata.URL))
	}

	// Footer
	s.WriteString("-----------------------------------\n")
	s.WriteString("*Happy Cooking! 👩‍🍳✨*\n")

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
