package models

import (
	"fmt"
	"strings"
	"time"
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

func formatRecipeContent(recipe *RecipeRaw) string {
	var s strings.Builder

	// Metadata
	s.WriteString(styles.headerStyle.Render("📝 Recipe Details"))
	s.WriteString("\n\n")
	s.WriteString(fmt.Sprintf("👤 Author: %s\n", recipe.Author))
	s.WriteString(fmt.Sprintf("⏲️ Cook Time: %v\n", recipe.CookTime))
	s.WriteString(fmt.Sprintf("📖 Description: %s\n", recipe.Description))
	s.WriteString("\n")

	// Ingredients
	s.WriteString(styles.headerStyle.Render("📋 Ingredients"))
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
		s.WriteString(styles.ingredientStyle.Render(ingredient.String()) + "\n")
	}
	s.WriteString("\n")

	// Instructions
	s.WriteString(styles.headerStyle.Render("🔨 Instructions"))
	s.WriteString("\n\n")
	for i, inst := range recipe.Instructions {
		// Add each instruction with proper padding and a newline
		s.WriteString(styles.instructionStyle.Render(fmt.Sprintf("%d. %s", i+1, inst)) + "\n")
	}
	s.WriteString("\n")

	// Categories
	if len(recipe.Categories) > 0 {
		s.WriteString(styles.headerStyle.Render("🏷️  Categories"))
		s.WriteString("\n\n")
		for _, cat := range recipe.Categories {
			s.WriteString(fmt.Sprintf("• %s\n", cat))
		}
	}

	return s.String()
}
