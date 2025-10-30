package utils

import (
	"fmt"
	"log/slog"
	"regexp"
	"strings"
)

type Ingredient struct {
	Amount  string
	Unit    string
	Name    string
	Details string
}

func ParseIngredient(input string) (Ingredient, error) {
	ingredient := Ingredient{}
	unitsPattern := strings.Join(CorpusMeasures, "|")
	re := regexp.MustCompile(fmt.Sprintf(`^(?:(\d+(?:/\d+)?(?:\s*-\s*\d+(?:/\d+)?)?\s*))?((?:%s)\s+)?([^(]+?)(?:\s*\((.*?)\))?$`, unitsPattern))
	matches := re.FindStringSubmatch(strings.ToLower(input))
	if len(matches) == 0 {
		return ingredient, fmt.Errorf("invalid ingredient")
	} else if len(matches) < 5 {
		ingredient.Name = strings.TrimSpace(input)
		return ingredient, nil
	}

	unit := strings.TrimSpace(matches[2])
	if unit != "" {
		if normalizedUnit, exists := CorpusMeasuresMap[unit]; exists {
			unit = normalizedUnit
		} else {
			unit = ""
		}
	}

	ingredient.Amount = strings.TrimSpace(matches[1])
	ingredient.Unit = unit
	ingredient.Name = strings.TrimSpace(matches[3])
	ingredient.Details = strings.TrimSpace(matches[4])
	return ingredient, nil
}

// parseIngredients extracts ingredients from the markdown
func ParseIngredientsFromMarkdown(text string) ([]Ingredient, error) {
	// Find the ingredients section
	ingredientsStart := strings.Index(text, "## 🥘 Ingredients")
	if ingredientsStart == -1 {
		return []Ingredient{}, fmt.Errorf("no ingredients found")
	}

	// Find the end of ingredients section (next ## or end of text)
	ingredientsEnd := strings.Index(text[ingredientsStart:], "\n## ")
	if ingredientsEnd == -1 {
		ingredientsEnd = len(text)
	} else {
		ingredientsEnd += ingredientsStart
	}

	ingredientsSection := text[ingredientsStart:ingredientsEnd]
	ingredients := []Ingredient{}
	lines := strings.Split(ingredientsSection, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "• ") {
			ingredientText := strings.TrimPrefix(line, "• ")
			ingredient, err := ParseIngredient(ingredientText)
			if err != nil {
				slog.Error("Failed to parse ingredient", "error", err)
				continue
			}
			ingredients = append(ingredients, ingredient)
		}
	}

	return ingredients, nil
}
