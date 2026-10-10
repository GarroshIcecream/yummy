package tools

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"time"

	"github.com/GarroshIcecream/yummy/internal/db"
	"github.com/GarroshIcecream/yummy/internal/utils"
	"github.com/tmc/langchaingo/tools"
)

// ListRecipesTool browses the cookbook with filters, sorting and pagination.
type ListRecipesTool struct {
	Cookbook *db.CookBook
}

var _ tools.Tool = &ListRecipesTool{}

const defaultPageSize = 10

func NewListRecipesTool(cookbook *db.CookBook) *ListRecipesTool {
	return &ListRecipesTool{Cookbook: cookbook}
}

func (t *ListRecipesTool) Name() string { return "listRecipes" }

func (t *ListRecipesTool) Description() string {
	return "List recipes in the cookbook, with optional filters, sorting and pagination. " +
		"Input: key=value pairs separated by ';' (all optional), e.g. " +
		"category=dessert; max_minutes=45; sort=quickest; page=2. " +
		"Keys: category, author, ingredient (comma-separated, all required), exclude (ingredients to avoid), " +
		"max_minutes (total time), min_rating (1-5), favourite=true, " +
		"sort (name | newest | rating | quickest), page, page_size (max 25). " +
		"Leave the input empty to list everything."
}

func (t *ListRecipesTool) Call(ctx context.Context, input string) (string, error) {
	slog.Debug("Executing tool", "tool", t.Name(), "input", input)
	args := parseArgs(input)
	if q := args["query"]; q != "" && len(args) == 1 {
		// A bare value is most likely a category ("desserts").
		args = map[string]string{"category": q}
	}

	recipes, err := fullRecipes(t.Cookbook)
	if err != nil {
		return "", err
	}

	category := strings.ToLower(args["category"])
	author := strings.ToLower(args["author"])
	include := listArg(args, "ingredient")
	include = append(include, listArg(args, "ingredients")...)
	exclude := listArg(args, "exclude")
	maxMinutes := intArg(args, "max_minutes", 0)
	minRating := intArg(args, "min_rating", 0)
	favourite := boolArg(args, "favourite") || boolArg(args, "favorite")

	var filtered []utils.RecipeRaw
	for _, r := range recipes {
		if category != "" && !hasCategory(r, category) {
			continue
		}
		if author != "" && !containsFold(r.Metadata.Author, author) {
			continue
		}
		if maxMinutes > 0 && (r.Metadata.TotalTime <= 0 || r.Metadata.TotalTime > time.Duration(maxMinutes)*time.Minute) {
			continue
		}
		if minRating > 0 && int(r.Metadata.Rating) < minRating {
			continue
		}
		if favourite && !r.IsFavourite && !r.Metadata.Favourite {
			continue
		}
		if !allIngredients(r, include) || anyIngredient(r, exclude) {
			continue
		}
		filtered = append(filtered, r)
	}

	sortRecipes(filtered, strings.ToLower(args["sort"]))

	pageSize := min(intArg(args, "page_size", defaultPageSize), 25)
	page := intArg(args, "page", 1)
	start, end, pages := paginate(len(filtered), page, pageSize)
	page = min(page, pages)

	var b strings.Builder
	if len(filtered) == 0 {
		b.WriteString("No recipes match those filters.")
		if cats, err := t.Cookbook.GetAllCategories(); err == nil && len(cats) > 0 {
			fmt.Fprintf(&b, " Known categories: %s.", strings.Join(cats, ", "))
		}
		return b.String(), nil
	}
	fmt.Fprintf(&b, "%d recipe(s), page %d of %d:\n", len(filtered), page, pages)
	for _, r := range filtered[start:end] {
		b.WriteString(summaryLine(r) + "\n")
	}
	if page < pages {
		fmt.Fprintf(&b, "\nMore available: call listRecipes again with page=%d.", page+1)
	}
	return b.String(), nil
}

func hasCategory(r utils.RecipeRaw, category string) bool {
	category = strings.TrimSuffix(category, "s") // "desserts" matches "Dessert"
	for _, c := range r.Metadata.Categories {
		if containsFold(c, category) {
			return true
		}
	}
	return false
}

func allIngredients(r utils.RecipeRaw, terms []string) bool {
	for _, term := range terms {
		if !ingredientMatches(r, term) {
			return false
		}
	}
	return true
}

func anyIngredient(r utils.RecipeRaw, terms []string) bool {
	for _, term := range terms {
		if ingredientMatches(r, term) {
			return true
		}
	}
	return false
}

func sortRecipes(recipes []utils.RecipeRaw, by string) {
	switch by {
	case "newest", "recent":
		sort.SliceStable(recipes, func(i, j int) bool {
			return recipes[i].Metadata.CreatedAt.After(recipes[j].Metadata.CreatedAt)
		})
	case "rating", "best":
		sort.SliceStable(recipes, func(i, j int) bool {
			return recipes[i].Metadata.Rating > recipes[j].Metadata.Rating
		})
	case "quickest", "fastest", "time":
		sort.SliceStable(recipes, func(i, j int) bool {
			ti, tj := recipes[i].Metadata.TotalTime, recipes[j].Metadata.TotalTime
			if ti <= 0 {
				return false // unknown times last
			}
			return tj <= 0 || ti < tj
		})
	default:
		sort.SliceStable(recipes, func(i, j int) bool {
			return strings.ToLower(recipes[i].RecipeName) < strings.ToLower(recipes[j].RecipeName)
		})
	}
}
