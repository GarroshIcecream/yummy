package recipe

import (
	"fmt"
	"strings"
	"time"
)

type RecipeWithDescription struct {
	RecipeID             uint
	RecipeName           string
	FormattedDescription string
}

func (i RecipeWithDescription) Title() string       { return i.RecipeName }
func (i RecipeWithDescription) Description() string { return i.FormattedDescription }
func (i RecipeWithDescription) FilterValue() string {
	return fmt.Sprintf("%s - %s", i.RecipeName, i.FormattedDescription)
}

func FormatRecipe(
	id uint,
	name string,
	author string,
	description string) RecipeWithDescription {

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
	}

}

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

	// Description in a styled blockquote
	if recipe.Description != "" {
		s.WriteString("💭 *About this recipe:*\n")
		s.WriteString(fmt.Sprintf("> %s\n\n", recipe.Description))
	}

	// Kitchen Preparation Box
	s.WriteString("## 👩‍🍳 Kitchen Prep\n\n")
	s.WriteString("| Timing & Portions | Details |\n")
	s.WriteString("|-------------------|----------|\n")
	if recipe.Author != "" {
		s.WriteString(fmt.Sprintf("| 👨‍🍳 Recipe By | *%s* |\n", recipe.Author))
	}
	if recipe.Quantity != "" {
		s.WriteString(fmt.Sprintf("| 🍽️ Servings | **%s** |\n", recipe.Quantity))
	}
	if recipe.TotalTime > 0 {
		s.WriteString(fmt.Sprintf("| ⏱️ Total Time | **%v** |\n", recipe.TotalTime))
	}
	if recipe.PrepTime > 0 {
		s.WriteString(fmt.Sprintf("| 🔪 Prep Time | **%v** |\n", recipe.PrepTime))
	}
	if recipe.CookTime > 0 {
		s.WriteString(fmt.Sprintf("| 🔥 Cook Time | **%v** |\n", recipe.CookTime))
	}
	s.WriteString("\n")

	// Ingredients section with better organization
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
