package list

import (
	"strings"

	db "github.com/GarroshIcecream/yummy/yummy/db"
)

// generateFilterSuggestions generates autocomplete suggestions based on the query
func generateFilterSuggestions(query string, cookbook *db.CookBook) []string {
	query = strings.TrimSpace(query)
	queryLower := strings.ToLower(query)
	var suggestions []string
	maxItems := 10

	if query == "" {
		// Show filter commands when empty
		return []string{
			"@author ",
			"@category ",
			"@ingredients ",
			"@description ",
			"@url ",
			"@fav",
		}
	}

	// Check if it's a filter command
	if strings.HasPrefix(query, "@") {
		// Handle filter commands
		if strings.HasPrefix(query, "@author ") {
			authorPrefix := strings.TrimSpace(strings.TrimPrefix(query, "@author "))
			authors, err := cookbook.GetAllAuthors()
			if err == nil {
				for _, author := range authors {
					if strings.HasPrefix(strings.ToLower(author), strings.ToLower(authorPrefix)) {
						suggestions = append(suggestions, "@author "+author)
						if len(suggestions) >= maxItems {
							break
						}
					}
				}
			}
		} else if strings.HasPrefix(query, "@category ") {
			categoryPrefix := strings.TrimSpace(strings.TrimPrefix(query, "@category "))
			categories, err := cookbook.GetAllCategories()
			if err == nil {
				for _, category := range categories {
					if strings.HasPrefix(strings.ToLower(category), strings.ToLower(categoryPrefix)) {
						suggestions = append(suggestions, "@category "+category)
						if len(suggestions) >= maxItems {
							break
						}
					}
				}
			}
		} else if strings.HasPrefix(query, "@ingredients ") {
			ingredientPrefix := strings.TrimSpace(strings.TrimPrefix(query, "@ingredients "))
			recipes, err := cookbook.AllRecipes()
			if err == nil {
				ingredientMap := make(map[string]bool)
				for _, recipe := range recipes {
					for _, ing := range recipe.Metadata.Ingredients {
						ingName := strings.ToLower(ing.Name)
						if strings.HasPrefix(ingName, strings.ToLower(ingredientPrefix)) {
							if !ingredientMap[ing.Name] {
								ingredientMap[ing.Name] = true
								suggestions = append(suggestions, "@ingredients "+ing.Name)
								if len(suggestions) >= maxItems {
									break
								}
							}
						}
					}
					if len(suggestions) >= maxItems {
						break
					}
				}
			}
		} else if strings.HasPrefix(query, "@description ") {
			descPrefix := strings.TrimSpace(strings.TrimPrefix(query, "@description "))
			recipes, err := cookbook.AllRecipes()
			if err == nil {
				descMap := make(map[string]bool)
				for _, recipe := range recipes {
					if recipe.RecipeDescription != "" {
						desc := strings.ToLower(recipe.RecipeDescription)
						if strings.Contains(desc, strings.ToLower(descPrefix)) {
							shortDesc := recipe.RecipeDescription
							if len(shortDesc) > 50 {
								shortDesc = shortDesc[:47] + "..."
							}
							key := "@description " + shortDesc
							if !descMap[key] && len(suggestions) < maxItems {
								descMap[key] = true
								suggestions = append(suggestions, key)
							}
						}
					}
				}
			}
		} else if strings.HasPrefix(query, "@url ") {
			urlPrefix := strings.TrimSpace(strings.TrimPrefix(query, "@url "))
			recipes, err := cookbook.AllRecipes()
			if err == nil {
				urlMap := make(map[string]bool)
				for _, recipe := range recipes {
					if recipe.Metadata.URL != "" {
						url := strings.ToLower(recipe.Metadata.URL)
						if strings.Contains(url, strings.ToLower(urlPrefix)) {
							if !urlMap[recipe.Metadata.URL] {
								urlMap[recipe.Metadata.URL] = true
								suggestions = append(suggestions, "@url "+recipe.Metadata.URL)
								if len(suggestions) >= maxItems {
									break
								}
							}
						}
					}
				}
			}
		} else {
			// Suggest filter commands
			filterCommands := []string{
				"@author ",
				"@category ",
				"@ingredients ",
				"@description ",
				"@url ",
				"@fav",
			}
			for _, cmd := range filterCommands {
				if strings.HasPrefix(cmd, queryLower) {
					suggestions = append(suggestions, cmd)
				}
			}
		}
	} else {
		// Regular recipe name search - suggest recipe names
		recipes, err := cookbook.AllRecipes()
		if err == nil {
			for _, recipe := range recipes {
				recipeName := recipe.RecipeName
				if strings.HasPrefix(strings.ToLower(recipeName), queryLower) {
					suggestions = append(suggestions, recipeName)
					if len(suggestions) >= maxItems {
						break
					}
				}
			}
		}
	}

	return suggestions
}
