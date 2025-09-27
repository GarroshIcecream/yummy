package recipe

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type RecipeTableItem struct {
	title  string
	value  string
	length int
}

type RecipeWithDescription struct {
	RecipeID             uint
	RecipeName           string
	FormattedDescription string
	IsFavourite          bool
}

func (i RecipeWithDescription) Title() string {
	if i.IsFavourite {
		return "⭐ " + i.RecipeName
	}
	return i.RecipeName
}
func (i RecipeWithDescription) Description() string { return i.FormattedDescription }
func (i RecipeWithDescription) FilterValue() string {
	return fmt.Sprintf("%s - %s", i.RecipeName, i.FormattedDescription)
}

type RecipeRaw struct {
	ID           uint
	Name         string
	Description  string
	Author       string
	CookTime     time.Duration
	PrepTime     time.Duration
	TotalTime    time.Duration
	Quantity     string
	URL          string
	Ingredients  []Ingredient
	Categories   []string
	Instructions []string
}

func FormatRecipe(
	id uint,
	name string,
	author string,
	description string,
	isFavourite bool) RecipeWithDescription {

	author_fin := "N/A"
	if author != "" {
		author_fin = author
	}

	desc := "N/A"
	if description != "" {
		desc = description
	}

	return RecipeWithDescription{
		RecipeID:             id,
		RecipeName:           name,
		FormattedDescription: fmt.Sprintf("%s - %s", author_fin, desc),
		IsFavourite:          isFavourite,
	}

}

func ConstructTableRow(item *RecipeTableItem, first_column_length int, longest_string int) string {
	r_pad_value := longest_string - item.length
	r_pad_title := first_column_length - len(item.title)
	return fmt.Sprintf("| %s | %s |\n", item.title+strings.Repeat(" ", r_pad_title), item.value+strings.Repeat(" ", r_pad_value))
}

func FormatRecipeContent(recipe *RecipeRaw) string {
	var s strings.Builder

	s.WriteString(fmt.Sprintf("# 🍳 %s\n\n", recipe.Name))

	if recipe.Description != "" {
		s.WriteString("💭 *About this recipe:*\n")
		s.WriteString(fmt.Sprintf("> %s\n\n", recipe.Description))
	}

	s.WriteString("## 👩‍🍳 Kitchen Prep\n\n")
	recipe_table := []RecipeTableItem{
		{title: "👨‍🍳 Recipe By", value: recipe.Author, length: len(recipe.Author)},
		{title: "🍽️ Servings", value: recipe.Quantity, length: len(recipe.Quantity)},
		{title: "⏱️ Total Time", value: recipe.TotalTime.String(), length: len(recipe.TotalTime.String())},
		{title: "🔪 Prep Time", value: recipe.PrepTime.String(), length: len(recipe.PrepTime.String())},
		{title: "🔥 Cook Time", value: recipe.CookTime.String(), length: len(recipe.CookTime.String())},
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

	for _, ing := range recipe.Ingredients {
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

	for i, inst := range recipe.Instructions {
		s.WriteString(fmt.Sprintf("%d. %s\n\n", i+1, inst))
	}

	// Categories and Tags
	if len(recipe.Categories) > 0 {
		s.WriteString("## 🏷️ Recipe Type\n\n")
		s.WriteString("*This recipe falls under:*\n\n")
		for _, cat := range recipe.Categories {
			s.WriteString(fmt.Sprintf("`%s` ", cat))
		}
		s.WriteString("\n\n")
	}

	// Source Attribution
	if recipe.URL != "" {
		s.WriteString("## 📖 Recipe Source\n\n")
		s.WriteString("*Want to learn more? Check out the original recipe:*\n\n")
		s.WriteString(fmt.Sprintf("🔗 [View Original Recipe](%s)\n\n", recipe.URL))
	}

	// Footer
	s.WriteString("-----------------------------------\n")
	s.WriteString("*Happy Cooking! 👩‍🍳✨*\n")
	s.WriteString("-----------------------------------\n")

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
		Ingredients:  []Ingredient{},
		Instructions: []string{},
		Categories:   []string{},
	}

	// Parse recipe name
	if customName != "" {
		recipeData.Name = customName
	} else {
		nameMatch := regexp.MustCompile(`(?m)^# 🍳 (.+)$`).FindStringSubmatch(text)
		if len(nameMatch) > 1 {
			recipeData.Name = strings.TrimSpace(nameMatch[1])
		} else {
			baseName := filepath.Base(filePath)
			recipeData.Name = strings.TrimSuffix(baseName, filepath.Ext(baseName))
		}
	}

	// Parse description
	descMatch := regexp.MustCompile(`💭 \*About this recipe:\*\n> (.+?)\n`).FindStringSubmatch(text)
	if len(descMatch) > 1 {
		recipeData.Description = strings.TrimSpace(descMatch[1])
	}

	// Parse metadata table
	parseMetadataTable(text, recipeData)

	// Parse ingredients
	parseIngredients(text, recipeData)

	// Parse instructions
	parseInstructions(text, recipeData)

	// Parse categories
	parseCategories(text, recipeData)

	// Parse source URL
	parseSourceURL(text, recipeData)

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
		Name:         jsonRecipe.Name,
		Description:  jsonRecipe.Description,
		Author:       jsonRecipe.Author,
		Quantity:     jsonRecipe.Quantity,
		URL:          jsonRecipe.URL,
		Ingredients:  []Ingredient{},
		Instructions: jsonRecipe.Instructions,
		Categories:   jsonRecipe.Categories,
	}

	// Override name if custom name provided
	if customName != "" {
		recipeData.Name = customName
	}

	// Parse time durations
	recipeData.CookTime = parseDuration(jsonRecipe.CookTime)
	recipeData.PrepTime = parseDuration(jsonRecipe.PrepTime)
	recipeData.TotalTime = parseDuration(jsonRecipe.TotalTime)

	// Convert ingredients
	for _, ing := range jsonRecipe.Ingredients {
		recipeData.Ingredients = append(recipeData.Ingredients, Ingredient{
			Amount:  ing.Amount,
			Unit:    ing.Unit,
			Name:    ing.Name,
			Details: ing.Details,
		})
	}

	return recipeData, nil
}

// parseMetadataTable extracts metadata from the markdown table
func parseMetadataTable(text string, recipeData *RecipeRaw) {
	// Parse author
	authorMatch := regexp.MustCompile(`👨‍🍳 Recipe By\s*\|\s*(.+?)\s*\|`).FindStringSubmatch(text)
	if len(authorMatch) > 1 {
		recipeData.Author = strings.TrimSpace(authorMatch[1])
	}

	// Parse servings
	servingsMatch := regexp.MustCompile(`🍽️ Servings\s*\|\s*(.+?)\s*\|`).FindStringSubmatch(text)
	if len(servingsMatch) > 1 {
		recipeData.Quantity = strings.TrimSpace(servingsMatch[1])
	}

	// Parse times
	totalTimeMatch := regexp.MustCompile(`⏱️ Total Time\s*\|\s*(.+?)\s*\|`).FindStringSubmatch(text)
	if len(totalTimeMatch) > 1 {
		recipeData.TotalTime = parseDuration(strings.TrimSpace(totalTimeMatch[1]))
	}

	prepTimeMatch := regexp.MustCompile(`🔪 Prep Time\s*\|\s*(.+?)\s*\|`).FindStringSubmatch(text)
	if len(prepTimeMatch) > 1 {
		recipeData.PrepTime = parseDuration(strings.TrimSpace(prepTimeMatch[1]))
	}

	cookTimeMatch := regexp.MustCompile(`🔥 Cook Time\s*\|\s*(.+?)\s*\|`).FindStringSubmatch(text)
	if len(cookTimeMatch) > 1 {
		recipeData.CookTime = parseDuration(strings.TrimSpace(cookTimeMatch[1]))
	}
}

// parseIngredients extracts ingredients from the markdown
func parseIngredients(text string, recipeData *RecipeRaw) {
	// Find the ingredients section
	ingredientsStart := strings.Index(text, "## 🥘 Ingredients")
	if ingredientsStart == -1 {
		return
	}

	// Find the end of ingredients section (next ## or end of text)
	ingredientsEnd := strings.Index(text[ingredientsStart:], "\n## ")
	if ingredientsEnd == -1 {
		ingredientsEnd = len(text)
	} else {
		ingredientsEnd += ingredientsStart
	}

	ingredientsSection := text[ingredientsStart:ingredientsEnd]

	// Parse each ingredient line
	lines := strings.Split(ingredientsSection, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "• ") {
			ingredientText := strings.TrimPrefix(line, "• ")
			ingredient := ParseIngredient(ingredientText)
			recipeData.Ingredients = append(recipeData.Ingredients, ingredient)
		}
	}
}

// parseInstructions extracts instructions from the markdown
func parseInstructions(text string, recipeData *RecipeRaw) {
	// Find the instructions section
	instructionsStart := strings.Index(text, "## 👩‍🍳 Cooking Instructions")
	if instructionsStart == -1 {
		return
	}

	// Find the end of instructions section
	instructionsEnd := strings.Index(text[instructionsStart:], "\n## ")
	if instructionsEnd == -1 {
		instructionsEnd = len(text)
	} else {
		instructionsEnd += instructionsStart
	}

	instructionsSection := text[instructionsStart:instructionsEnd]

	// Parse numbered instructions
	lines := strings.Split(instructionsSection, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Match numbered list items (1. 2. etc.)
		if match := regexp.MustCompile(`^\d+\.\s*(.+)$`).FindStringSubmatch(line); len(match) > 1 {
			recipeData.Instructions = append(recipeData.Instructions, strings.TrimSpace(match[1]))
		}
	}
}

// parseCategories extracts categories from the markdown
func parseCategories(text string, recipeData *RecipeRaw) {
	// Find the categories section
	categoriesStart := strings.Index(text, "## 🏷️ Recipe Type")
	if categoriesStart == -1 {
		return
	}

	// Find the end of categories section
	categoriesEnd := strings.Index(text[categoriesStart:], "\n## ")
	if categoriesEnd == -1 {
		categoriesEnd = len(text)
	} else {
		categoriesEnd += categoriesStart
	}

	categoriesSection := text[categoriesStart:categoriesEnd]

	// Extract categories from backticks
	re := regexp.MustCompile("`([^`]+)`")
	matches := re.FindAllStringSubmatch(categoriesSection, -1)
	for _, match := range matches {
		if len(match) > 1 {
			recipeData.Categories = append(recipeData.Categories, strings.TrimSpace(match[1]))
		}
	}
}

// parseSourceURL extracts the source URL from the markdown
func parseSourceURL(text string, recipeData *RecipeRaw) {
	urlMatch := regexp.MustCompile(`🔗 \[View Original Recipe\]\((.+?)\)`).FindStringSubmatch(text)
	if len(urlMatch) > 1 {
		recipeData.URL = strings.TrimSpace(urlMatch[1])
	}
}

// parseDuration parses a duration string into time.Duration
func parseDuration(durationStr string) time.Duration {
	if durationStr == "" || durationStr == "N/A" {
		return 0
	}

	// Handle common duration formats
	durationStr = strings.ToLower(strings.TrimSpace(durationStr))

	// Try to parse as Go duration first
	if duration, err := time.ParseDuration(durationStr); err == nil {
		return duration
	}

	// Handle "X hours Y minutes" format
	if match := regexp.MustCompile(`(\d+)\s*hours?\s*(\d+)\s*minutes?`).FindStringSubmatch(durationStr); len(match) > 2 {
		hours, _ := strconv.Atoi(match[1])
		minutes, _ := strconv.Atoi(match[2])
		return time.Duration(hours)*time.Hour + time.Duration(minutes)*time.Minute
	}

	// Handle "X hours" format
	if match := regexp.MustCompile(`(\d+)\s*hours?`).FindStringSubmatch(durationStr); len(match) > 1 {
		hours, _ := strconv.Atoi(match[1])
		return time.Duration(hours) * time.Hour
	}

	// Handle "X minutes" format
	if match := regexp.MustCompile(`(\d+)\s*minutes?`).FindStringSubmatch(durationStr); len(match) > 1 {
		minutes, _ := strconv.Atoi(match[1])
		return time.Duration(minutes) * time.Minute
	}

	return 0
}
