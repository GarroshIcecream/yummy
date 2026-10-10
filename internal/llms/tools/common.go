package tools

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/GarroshIcecream/yummy/internal/db"
	"github.com/GarroshIcecream/yummy/internal/utils"
)

// The agent passes tool input as free text, and small local models are not
// reliable at producing strict JSON. parseArgs accepts all of:
//
//	{"category": "dessert", "page": 2}     JSON object
//	category=dessert, page=2               key=value pairs (',' or ';')
//	dessert                                a bare value (returned under "query")
//
// Keys are lower-cased; surrounding quotes and backticks are stripped.
func parseArgs(input string) map[string]string {
	args := map[string]string{}
	input = strings.TrimSpace(input)
	input = strings.Trim(input, "`")
	input = strings.TrimSpace(input)
	if input == "" {
		return args
	}

	if strings.HasPrefix(input, "{") {
		var raw map[string]any
		if err := json.Unmarshal([]byte(input), &raw); err == nil {
			for k, v := range raw {
				switch val := v.(type) {
				case string:
					args[strings.ToLower(k)] = strings.TrimSpace(val)
				case []any:
					parts := make([]string, 0, len(val))
					for _, p := range val {
						parts = append(parts, fmt.Sprint(p))
					}
					args[strings.ToLower(k)] = strings.Join(parts, ",")
				case nil:
				default:
					args[strings.ToLower(k)] = fmt.Sprint(val)
				}
			}
			return args
		}
	}

	if strings.Contains(input, "=") {
		// key=value pairs separated by , ; & or newlines. A piece without "="
		// continues the previous value, so lists like
		// "ingredients=chicken, rice" survive the comma split.
		last := ""
		for _, piece := range strings.FieldsFunc(input, func(r rune) bool {
			return r == ',' || r == ';' || r == '&' || r == '\n'
		}) {
			if k, v, ok := strings.Cut(piece, "="); ok {
				last = strings.ToLower(strings.TrimSpace(k))
				args[last] = unquote(v)
			} else if last != "" {
				args[last] += "," + unquote(piece)
			}
		}
		if len(args) > 0 {
			return args
		}
	}

	args["query"] = unquote(input)
	return args
}

func unquote(s string) string {
	s = strings.TrimSpace(s)
	return strings.TrimSpace(strings.Trim(s, `"'`))
}

// intArg reads a positive integer argument, falling back to def.
func intArg(args map[string]string, key string, def int) int {
	if v, ok := args[key]; ok {
		if n, err := strconv.Atoi(strings.TrimSpace(v)); err == nil && n > 0 {
			return n
		}
	}
	return def
}

func boolArg(args map[string]string, key string) bool {
	switch strings.ToLower(args[key]) {
	case "true", "yes", "1", "y":
		return true
	}
	return false
}

// listArg splits a comma-separated argument into trimmed, lower-cased items.
func listArg(args map[string]string, key string) []string {
	var out []string
	for p := range strings.SplitSeq(args[key], ",") {
		if p = strings.ToLower(unquote(p)); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// fullRecipes loads every recipe with ingredients, instructions and rating.
// The cookbook is a local SQLite database of personal recipes, so loading it
// in full per tool call is cheap and keeps filtering logic in one place.
func fullRecipes(cookbook *db.CookBook) ([]utils.RecipeRaw, error) {
	summaries, err := cookbook.AllRecipes()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch recipes: %w", err)
	}
	recipes := make([]utils.RecipeRaw, 0, len(summaries))
	for _, s := range summaries {
		full, err := cookbook.GetFullRecipe(s.RecipeID)
		if err != nil || full == nil {
			recipes = append(recipes, s)
			continue
		}
		full.IsFavourite = s.IsFavourite
		recipes = append(recipes, *full)
	}
	return recipes, nil
}

// summaryLine renders one compact result line for list-style tool output.
func summaryLine(r utils.RecipeRaw) string {
	var facts []string
	if t := utils.FormatDurationHuman(r.Metadata.TotalTime); t != "" {
		facts = append(facts, t)
	}
	if r.Metadata.Rating > 0 {
		facts = append(facts, fmt.Sprintf("rated %d/5", r.Metadata.Rating))
	}
	if r.IsFavourite || r.Metadata.Favourite {
		facts = append(facts, "favourite")
	}
	if len(r.Metadata.Categories) > 0 {
		facts = append(facts, strings.Join(r.Metadata.Categories, ", "))
	}
	if r.Metadata.Author != "" {
		facts = append(facts, "by "+r.Metadata.Author)
	}
	line := fmt.Sprintf("- [ID %d] %s", r.RecipeID, r.RecipeName)
	if len(facts) > 0 {
		line += " (" + strings.Join(facts, " | ") + ")"
	}
	return line
}

// paginate returns the page slice bounds and the page count.
func paginate(total, page, pageSize int) (start, end, pages int) {
	pages = max(1, (total+pageSize-1)/pageSize)
	page = min(max(page, 1), pages)
	start = (page - 1) * pageSize
	end = min(start+pageSize, total)
	return start, end, pages
}

func containsFold(haystack, needle string) bool {
	return strings.Contains(strings.ToLower(haystack), strings.ToLower(needle))
}

func ingredientMatches(r utils.RecipeRaw, term string) bool {
	for _, ing := range r.Metadata.Ingredients {
		if containsFold(ing.Name, term) || containsFold(ing.BaseName, term) {
			return true
		}
	}
	return false
}
