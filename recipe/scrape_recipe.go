package recipe

import (
	"time"

	scrape "github.com/kkyr/go-recipe"
	recipe "github.com/kkyr/go-recipe/pkg/recipe"
)

func SafeGet[T any](s scrape.Scraper, getter func(scrape.Scraper) (T, bool)) *T {
	if value, ok := getter(s); ok {
		return &value
	}
	return nil
}

func GetIngredients(s scrape.Scraper) *[]string {
	return SafeGet(s, scrape.Scraper.Ingredients)
}

func GetInstructions(s scrape.Scraper) *[]string {
	return SafeGet(s, scrape.Scraper.Instructions)
}

func GetAuthor(s scrape.Scraper) *string {
	return SafeGet(s, scrape.Scraper.Author)
}

func GetCookTime(s scrape.Scraper) *time.Duration {
	return SafeGet(s, scrape.Scraper.CookTime)
}

func GetDescription(s scrape.Scraper) *string {
	return SafeGet(s, scrape.Scraper.Description)
}

func GetQuantity(s scrape.Scraper) *string {
	return SafeGet(s, scrape.Scraper.Yields)
}

func GetCategories(s scrape.Scraper) *[]string {
	return SafeGet(s, scrape.Scraper.Categories)
}

func GetName(s scrape.Scraper) *string {
	return SafeGet(s, scrape.Scraper.Name)
}

func GetRecipeFromURL(url string) (*RecipeRaw, error) {
	recipe, err := recipe.ScrapeURL(url)
	if err != nil {
		return nil, err
	}

	recipeRaw := &RecipeRaw{
		URL: url,
	}

	if name := GetName(recipe); name != nil {
		recipeRaw.Name = *name
	}
	if desc := GetDescription(recipe); desc != nil {
		recipeRaw.Description = *desc
	}
	if author := GetAuthor(recipe); author != nil {
		recipeRaw.Author = *author
	}
	if cookTime := GetCookTime(recipe); cookTime != nil {
		recipeRaw.CookTime = *cookTime
	}
	if quantity := GetQuantity(recipe); quantity != nil {
		recipeRaw.Quantity = *quantity
	}
	if categories := GetCategories(recipe); categories != nil {
		recipeRaw.Categories = *categories
	}
	if instructions := GetInstructions(recipe); instructions != nil {
		recipeRaw.Instructions = *instructions
	}
	if ingredients := GetIngredients(recipe); ingredients != nil {
		parsedIngredients := make([]Ingredient, len(*ingredients))
		for i, ing := range *ingredients {
			parsedIngredients[i] = ParseIngredient(ing)
		}
		recipeRaw.Ingredients = parsedIngredients
	}

	return recipeRaw, nil
}
