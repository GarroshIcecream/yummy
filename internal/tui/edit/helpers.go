package edit

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	utils "github.com/GarroshIcecream/yummy/internal/utils"
)

var ingredientBulletPrefix = regexp.MustCompile(`^\s*[-*•]\s+`)
var instructionPrefix = regexp.MustCompile(`^\s*(?:[-*•]\s+|\d+[.)]\s+)`)

func formatDurationInput(d time.Duration) string {
	if d <= 0 {
		return ""
	}

	hours := int(d / time.Hour)
	minutes := int((d % time.Hour) / time.Minute)
	parts := make([]string, 0, 2)
	if hours > 0 {
		parts = append(parts, fmt.Sprintf("%dh", hours))
	}
	if minutes > 0 {
		parts = append(parts, fmt.Sprintf("%dm", minutes))
	}
	if len(parts) == 0 {
		parts = append(parts, "0m")
	}
	return strings.Join(parts, " ")
}

func parseDurationInput(raw string) (time.Duration, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return 0, nil
	}

	compact := strings.Join(strings.Fields(trimmed), "")
	if d, err := time.ParseDuration(compact); err == nil {
		return d, nil
	}

	parsed := utils.ParseDurationFromString(trimmed)
	if parsed > 0 {
		return parsed, nil
	}

	return 0, fmt.Errorf("must be a valid duration")
}

func validateDurationField(raw string) error {
	_, err := parseDurationInput(raw)
	return err
}

func validateOptionalURL(raw string) error {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	return utils.ValidateURL(strings.TrimSpace(raw))
}

func categoriesToText(categories []string) string {
	return strings.Join(categories, ", ")
}

func parseCategories(raw string) []string {
	parts := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == '\n'
	})

	seen := make(map[string]bool)
	categories := make([]string, 0, len(parts))
	for _, part := range parts {
		category := strings.TrimSpace(part)
		if category == "" {
			continue
		}
		key := strings.ToLower(category)
		if seen[key] {
			continue
		}
		seen[key] = true
		categories = append(categories, category)
	}
	return categories
}

func ingredientToEditableLine(ingredient utils.Ingredient) string {
	parts := make([]string, 0, 3)
	if ingredient.Amount != "" {
		parts = append(parts, ingredient.Amount)
	}
	if ingredient.Unit != "" {
		parts = append(parts, ingredient.Unit)
	}
	if ingredient.Name != "" {
		parts = append(parts, ingredient.Name)
	}

	line := strings.Join(parts, " ")
	if ingredient.Details != "" {
		line = strings.TrimSpace(line)
		if line != "" {
			line += fmt.Sprintf(" (%s)", ingredient.Details)
		} else {
			line = ingredient.Details
		}
	}
	return strings.TrimSpace(line)
}

func ingredientsToText(ingredients []utils.Ingredient) string {
	if len(ingredients) == 0 {
		return ""
	}

	lines := make([]string, 0, len(ingredients))
	currentGroup := ""
	for _, ingredient := range ingredients {
		if ingredient.Group != currentGroup {
			currentGroup = ingredient.Group
			if currentGroup != "" {
				lines = append(lines, currentGroup+":")
			}
		}
		lines = append(lines, ingredientToEditableLine(ingredient))
	}

	return strings.Join(lines, "\n")
}

func parseIngredientLine(raw string) (utils.Ingredient, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return utils.Ingredient{}, fmt.Errorf("ingredient cannot be empty")
	}

	unitsPattern := strings.Join(utils.CorpusMeasures, "|")
	re := regexp.MustCompile(fmt.Sprintf(`^(?:(\d+(?:\.\d+)?(?:/\d+(?:\.\d+)?)?(?:\s*-\s*\d+(?:\.\d+)?(?:/\d+(?:\.\d+)?)?)?\s*))?((?:%s)\s+)?([^(]+?)(?:\s*\((.*?)\))?$`, unitsPattern))
	matches := re.FindStringSubmatch(trimmed)
	if len(matches) < 5 {
		return utils.Ingredient{Name: trimmed, BaseName: trimmed}, nil
	}

	unit := strings.TrimSpace(matches[2])
	if unit != "" {
		if normalizedUnit, exists := utils.CorpusMeasuresMap[strings.ToLower(unit)]; exists {
			unit = normalizedUnit
		} else {
			unit = ""
		}
	}

	name := strings.TrimSpace(matches[3])
	if name == "" {
		name = trimmed
	}

	return utils.Ingredient{
		Amount:   strings.TrimSpace(matches[1]),
		Unit:     unit,
		Name:     name,
		Details:  strings.TrimSpace(matches[4]),
		BaseName: name,
	}, nil
}

func parseIngredientsText(raw string) ([]utils.Ingredient, error) {
	lines := strings.Split(raw, "\n")
	ingredients := make([]utils.Ingredient, 0, len(lines))
	currentGroup := ""

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		if !ingredientBulletPrefix.MatchString(trimmed) && strings.HasSuffix(trimmed, ":") {
			currentGroup = strings.TrimSpace(strings.TrimSuffix(trimmed, ":"))
			continue
		}

		cleaned := ingredientBulletPrefix.ReplaceAllString(trimmed, "")
		ingredient, err := parseIngredientLine(cleaned)
		if err != nil {
			return nil, err
		}
		ingredient.Group = currentGroup
		ingredients = append(ingredients, ingredient)
	}

	if len(ingredients) == 0 {
		return nil, fmt.Errorf("add at least one ingredient")
	}

	return ingredients, nil
}

func validateIngredientsField(raw string) error {
	_, err := parseIngredientsText(raw)
	return err
}

func instructionsToText(instructions []string) string {
	return strings.Join(instructions, "\n")
}

func parseInstructionsText(raw string) ([]string, error) {
	lines := strings.Split(raw, "\n")
	instructions := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		trimmed = instructionPrefix.ReplaceAllString(trimmed, "")
		trimmed = strings.TrimSpace(trimmed)
		if trimmed == "" {
			continue
		}
		instructions = append(instructions, trimmed)
	}

	if len(instructions) == 0 {
		return nil, fmt.Errorf("add at least one instruction")
	}

	return instructions, nil
}

func validateInstructionsField(raw string) error {
	_, err := parseInstructionsText(raw)
	return err
}
