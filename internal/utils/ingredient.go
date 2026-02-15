package utils

import (
	"fmt"
	"log/slog"
	"regexp"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type Ingredient struct {
	Amount   string
	Unit     string
	Name     string
	Details  string
	BaseName string // core ingredient word(s) for highlighting (e.g. "thyme" from "dried thyme")
}

func ParseIngredient(input string) (Ingredient, error) {
	ingredient := Ingredient{}
	unitsPattern := strings.Join(CorpusMeasures, "|")
	re := regexp.MustCompile(fmt.Sprintf(`^(?:(\d+(?:\.\d+)?(?:/\d+(?:\.\d+)?)?(?:\s*-\s*\d+(?:\.\d+)?(?:/\d+(?:\.\d+)?)?)?\s*))?((?:%s)\s+)?([^(]+?)(?:\s*\((.*?)\))?$`, unitsPattern))
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

// ingredientStopWords are common words that should NOT be treated as
// matchable ingredient tokens (too generic / likely to false-positive).
var ingredientStopWords = map[string]bool{
	"and": true, "or": true, "the": true, "for": true, "with": true,
	"into": true, "from": true, "each": true, "all": true, "fresh": true,
	"large": true, "small": true, "medium": true, "whole": true,
	"good": true, "fine": true, "extra": true, "light": true,
	"dark": true, "warm": true, "cold": true, "hot": true,
	"dry": true, "raw": true, "pure": true, "plain": true,
}

// extractIngredientTokens returns matchable tokens from ingredient names.
// When BaseName is populated (via LLM extraction) it is preferred as the
// primary token because it contains only the core ingredient word(s)
// (e.g. "thyme" instead of "dried thyme"). Full names and individual words
// are still added as secondary tokens. Tokens are sorted longest-first so
// "olive oil" is matched before "oil".
func extractIngredientTokens(ingredients []Ingredient) []string {
	seen := make(map[string]bool)
	var tokens []string

	addToken := func(t string) {
		t = strings.TrimSpace(t)
		if t == "" {
			return
		}
		key := strings.ToLower(t)
		if seen[key] || ingredientStopWords[key] {
			return
		}
		seen[key] = true
		tokens = append(tokens, t)
	}

	// First pass: add BaseName (preferred) or full Name as the primary token.
	for _, ing := range ingredients {
		if ing.BaseName != "" {
			addToken(ing.BaseName)
		} else {
			addToken(ing.Name)
		}
	}

	// Second pass: add full names that weren't already added via BaseName.
	for _, ing := range ingredients {
		addToken(ing.Name)
	}

	// Third pass: add individual words from base names and full names.
	for _, ing := range ingredients {
		// Prefer splitting BaseName if available.
		source := ing.BaseName
		if source == "" {
			source = ing.Name
		}
		words := strings.Fields(source)
		if len(words) <= 1 {
			continue
		}
		for _, w := range words {
			if len(w) >= 3 {
				addToken(w)
			}
		}
	}

	// Sort longest-first to prevent partial-match clobbering
	sort.Slice(tokens, func(i, j int) bool {
		return len(tokens[i]) > len(tokens[j])
	})
	return tokens
}

// buildIngredientPattern builds a compiled regex that matches any of the
// ingredient tokens (case-insensitive, word-boundary-aware).
// Returns nil if no valid tokens are produced.
func buildIngredientPattern(ingredients []Ingredient) *regexp.Regexp {
	tokens := extractIngredientTokens(ingredients)
	if len(tokens) == 0 {
		return nil
	}

	quoted := make([]string, len(tokens))
	for i, t := range tokens {
		quoted[i] = regexp.QuoteMeta(t)
	}
	pattern := `(?i)\b(` + strings.Join(quoted, "|") + `)\b`
	re, err := regexp.Compile(pattern)
	if err != nil {
		slog.Error("Failed to compile ingredient highlight regex", "error", err)
		return nil
	}
	return re
}

// HighlightIngredientsInMarkdown wraps ingredient names found in text with
// markdown bold markers for visual highlighting when rendered by glamour.
//
// Example: "Add the flour and sugar" → "Add the **flour** and **sugar**"
func HighlightIngredientsInMarkdown(text string, ingredients []Ingredient) string {
	re := buildIngredientPattern(ingredients)
	if re == nil {
		return text
	}
	return re.ReplaceAllString(text, "**$1**")
}

// HighlightIngredientsWithStyle returns text with ingredient names rendered
// using the provided lipgloss style, for use in direct terminal rendering
// (e.g. cooking mode).
func HighlightIngredientsWithStyle(text string, ingredients []Ingredient, style lipgloss.Style) string {
	re := buildIngredientPattern(ingredients)
	if re == nil {
		return text
	}
	return re.ReplaceAllStringFunc(text, func(match string) string {
		return style.Render(match)
	})
}

// ParseIngredientsFromMarkdown extracts ingredients from the markdown
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
