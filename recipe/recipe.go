package recipe

import (
	"fmt"
	"strings"
	"time"

	styles "github.com/GarroshIcecream/yummy/styles"
)

type RecipeRaw struct {
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

func FormatRecipeContent(recipe *RecipeRaw) string {
	var s strings.Builder

	s.WriteString(styles.HeaderStyle.Render(fmt.Sprintf("🎉 %s 🎉", recipe.Name)))
	s.WriteString("\n\n")

	// Metadata
	s.WriteString(styles.HeaderStyle.Render("📝 Recipe Details"))
	s.WriteString("\n\n")
	s.WriteString(fmt.Sprintf("👤 Author: %s\n", recipe.Author))
	s.WriteString(fmt.Sprintf("⏲️ Total Time: %v\n", recipe.TotalTime))
	s.WriteString(fmt.Sprintf("📖 Description: \n%s\n\n", recipe.Description))

	// Include URL if available
	if recipe.URL != "" {
		s.WriteString(fmt.Sprintf("🔗 URL: %s\n", recipe.URL))
	}
	s.WriteString("\n")

	// Ingredients
	s.WriteString(styles.HeaderStyle.Render("📋 Ingredients"))
	s.WriteString("\n\n")
	for _, ing := range recipe.Ingredients {
		var ingredient strings.Builder
		ingredient.WriteString("• ")

		if ing.Amount != "" {
			ingredient.WriteString(ing.Amount + " ")
		}

		if ing.Unit != "" {
			ingredient.WriteString(ing.Unit + " ")
		}

		ingredient.WriteString(ing.Name)

		if ing.Details != "" {
			ingredient.WriteString(fmt.Sprintf(" (%s)", ing.Details))
		}
		s.WriteString(styles.IngredientStyle.Render(ingredient.String()) + "\n")
	}
	s.WriteString("\n")

	// Instructions
	s.WriteString(styles.HeaderStyle.Render("🔨 Instructions"))
	s.WriteString("\n\n")
	for i, inst := range recipe.Instructions {
		// Add each instruction with proper padding and a newline
		s.WriteString(styles.InstructionStyle.Render(fmt.Sprintf("%d. %s", i+1, inst)) + "\n")
	}
	s.WriteString("\n")

	// Categories
	if len(recipe.Categories) > 0 {
		s.WriteString(styles.HeaderStyle.Render("🏷️  Categories"))
		s.WriteString("\n\n")
		for _, cat := range recipe.Categories {
			s.WriteString(fmt.Sprintf("• %s\n", cat))
		}
	}

	return s.String()
}
