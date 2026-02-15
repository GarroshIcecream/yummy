package scrape

import (
	"strings"
	"time"

	"github.com/kkyr/go-recipe"
)

type Scraper interface {
	Author() (string, bool)
	CanonicalURL() (string, bool)
	Host() (string, bool)
	Categories() ([]string, bool)
	CookTime() (time.Duration, bool)
	Cuisine() ([]string, bool)
	Description() (string, bool)
	ImageURL() (string, bool)
	Ingredients() ([]string, bool)
	Instructions() ([]string, bool)
	Language() (string, bool)
	Name() (string, bool)
	Nutrition() (recipe.Nutrition, bool)
	PrepTime() (time.Duration, bool)
	SuitableDiets() ([]recipe.Diet, bool)
	TotalTime() (time.Duration, bool)
	Yields() (string, bool)
}

type adapter struct {
	j RecipeScrapersJSON
}

func (a *adapter) Author() (string, bool) {
	return strings.TrimSpace(a.j.Author), a.j.Author != ""
}

func (a *adapter) CanonicalURL() (string, bool) {
	return strings.TrimSpace(a.j.CanonicalURL), a.j.CanonicalURL != ""
}

func (a *adapter) Host() (string, bool) {
	return strings.TrimSpace(a.j.Host), a.j.Host != ""
}

func (a *adapter) Categories() ([]string, bool) {
	if a.j.Category == "" {
		return nil, false
	}
	return []string{strings.TrimSpace(a.j.Category)}, true
}

func (a *adapter) CookTime() (time.Duration, bool) {
	if a.j.CookTime == nil || *a.j.CookTime <= 0 {
		return 0, false
	}
	return time.Duration(*a.j.CookTime) * time.Minute, true
}

func (a *adapter) Cuisine() ([]string, bool) {
	return nil, false
}

func (a *adapter) Description() (string, bool) {
	return strings.TrimSpace(a.j.Description), a.j.Description != ""
}

func (a *adapter) ImageURL() (string, bool) {
	return strings.TrimSpace(a.j.Image), a.j.Image != ""
}

func (a *adapter) Ingredients() ([]string, bool) {
	if len(a.j.Ingredients) == 0 {
		return nil, false
	}
	return a.j.Ingredients, true
}

func (a *adapter) Instructions() ([]string, bool) {
	if len(a.j.InstructionsList) > 0 {
		return a.j.InstructionsList, true
	}
	if a.j.Instructions == "" {
		return nil, false
	}
	steps := strings.Split(a.j.Instructions, "\n")
	var out []string
	for _, s := range steps {
		if t := strings.TrimSpace(s); t != "" {
			out = append(out, t)
		}
	}
	return out, len(out) > 0
}

func (a *adapter) Language() (string, bool) {
	return strings.TrimSpace(a.j.Language), a.j.Language != ""
}

func (a *adapter) Name() (string, bool) {
	return strings.TrimSpace(a.j.Title), a.j.Title != ""
}

func (a *adapter) Nutrition() (recipe.Nutrition, bool) {
	return recipe.Nutrition{}, false
}

func (a *adapter) PrepTime() (time.Duration, bool) {
	if a.j.PrepTime == nil || *a.j.PrepTime <= 0 {
		return 0, false
	}
	return time.Duration(*a.j.PrepTime) * time.Minute, true
}

func (a *adapter) SuitableDiets() ([]recipe.Diet, bool) {
	return nil, false
}

func (a *adapter) TotalTime() (time.Duration, bool) {
	if a.j.TotalTime == nil || *a.j.TotalTime <= 0 {
		return 0, false
	}
	return time.Duration(*a.j.TotalTime) * time.Minute, true
}

func (a *adapter) Yields() (string, bool) {
	return strings.TrimSpace(a.j.Yields), a.j.Yields != ""
}
