package db

import (
	"fmt"

	"recipe_me/models"
)

func main() {

	// _, err := db.NewCookBook()
	// if err != nil {
	// 	fmt.Println("Error creating CookBook:", err)
	// }

	// p := tea.NewProgram(models.NewModel())

	// if _, err := p.Run(); err != nil {
	// 	fmt.Printf("Alas, there's been an error: %v", err)
	// 	os.Exit(1)
	// }

	cookbook, err := NewCookBook()
	if err != nil {
		fmt.Println("Error creating CookBook:", err)
		return
	}

	err = cookbook.Open()
	if err != nil {
		fmt.Println("Error creating CookBook:", err)
		return
	}

	url := "https://blog.fatfreevegan.com/2013/06/kale-and-quinoa-salad-with-black-beans.html"

	// Get recipe from URL
	recipeRaw, err := models.GetRecipeFromURL(url)
	if err != nil {
		fmt.Printf("Error scraping recipe: %v\n", err)
		return
	}

	// Save to database
	err = cookbook.SaveScrapedRecipe(recipeRaw)
	if err != nil {
		fmt.Printf("Error saving recipe: %v\n", err)
		return
	}

	fmt.Println("Recipe saved successfully!")

	// Get and print the recipe
	// First, let's list all available recipes
	recipes, err := cookbook.AllRecipes()
	if err != nil {
		fmt.Printf("Error fetching recipes: %v\n", err)
		return
	}

	fmt.Println("Available recipes:")
	for _, r := range recipes {
		fmt.Printf("- %s (ID: %d)\n", r.RecipeName, r.ID)
	}

	recipeName := "Kale and Quinoa Salad with Black Beans" // or whatever name was saved
	recipe, err := cookbook.GetFullRecipe(recipeName)
	if err != nil {
		fmt.Printf("Error fetching recipe: %v\n", err)
		return
	}

	// Print the recipe details
	fmt.Printf("Recipe: %s\n", recipe.Name)
	fmt.Printf("Author: %s\n", recipe.Author)
	fmt.Printf("Description: %s\n", recipe.Description)
	fmt.Printf("Cook Time: %v\n", recipe.CookTime)
	fmt.Printf("URL: %s\n\n", recipe.URL)

	fmt.Println("Ingredients:")
	for _, ing := range recipe.Ingredients {
		if ing.Details != "" {
			fmt.Printf("- %s %s %s (%s)\n", ing.Amount, ing.Unit, ing.Name, ing.Details)
		} else {
			fmt.Printf("- %s %s %s\n", ing.Amount, ing.Unit, ing.Name)
		}
	}

	fmt.Println("\nInstructions:")
	for i, inst := range recipe.Instructions {
		fmt.Printf("%d. %s\n", i+1, inst)
	}

	if len(recipe.Categories) > 0 {
		fmt.Println("\nCategories:")
		for _, cat := range recipe.Categories {
			fmt.Printf("- %s\n", cat)
		}
	}
}
