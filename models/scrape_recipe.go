package models

import (
	scrape "github.com/kkyr/go-recipe"
	recipe "github.com/kkyr/go-recipe/pkg/recipe"
)

func get_recipe_from_url(url string) scrape.Scraper {

	recipe, err := recipe.ScrapeURL(url)
	if err != nil {
		// handle err
	}

	return recipe
}
