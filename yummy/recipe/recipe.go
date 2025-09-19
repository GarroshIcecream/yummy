package recipe

import (
	"fmt"
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
