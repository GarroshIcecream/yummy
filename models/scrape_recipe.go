package models

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

func GetName(s scrape.Scraper) *string {
	return SafeGet(s, scrape.Scraper.Name)
}

func get_recipe_from_url(url string) scrape.Scraper {

	recipe, err := recipe.ScrapeURL(url)
	if err != nil {
		// handle err
	}

	return recipe
}
