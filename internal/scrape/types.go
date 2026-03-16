package scrape

import (
	"encoding/json"
	"fmt"
	"strings"
)

// FlexibleString handles JSON values that may be either a string or an array of
// strings (e.g. keywords). When an array is encountered it is joined with ", ".
type FlexibleString string

func (f *FlexibleString) UnmarshalJSON(data []byte) error {
	// Try plain string first.
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*f = FlexibleString(s)
		return nil
	}

	// Try array of strings.
	var arr []string
	if err := json.Unmarshal(data, &arr); err == nil {
		*f = FlexibleString(strings.Join(arr, ", "))
		return nil
	}

	return fmt.Errorf("FlexibleString: expected string or []string, got %s", string(data))
}

// RecipeScrapersJSON is the JSON shape from recipe_scrapers.scrape_me(url).to_json().
type RecipeScrapersJSON struct {
	Author           string            `json:"author"`
	CanonicalURL     string            `json:"canonical_url"`
	Category         string            `json:"category"`
	CookTime         *int              `json:"cook_time"`
	Host             string            `json:"host"`
	Image            string            `json:"image"`
	IngredientGroups []IngredientGroup `json:"ingredient_groups"`
	Ingredients      []string          `json:"ingredients"`
	Instructions     string            `json:"instructions"`
	InstructionsList []string          `json:"instructions_list"`
	Keywords         FlexibleString    `json:"keywords"`
	Language         string            `json:"language"`
	Nutrients        map[string]any    `json:"nutrients"`
	PrepTime         *int              `json:"prep_time"`
	Title            string            `json:"title"`
	TotalTime        *int              `json:"total_time"`
	Description      string            `json:"description"`
	Yields           string            `json:"yields"`
}

type IngredientGroup struct {
	Ingredients []string `json:"ingredients"`
	Purpose     *string  `json:"purpose"`
}

// Nutrition represents nutritional information about a recipe.
type Nutrition struct {
	Calories              float32
	CarbohydrateGrams     float32
	CholesterolMilligrams float32
	FatGrams              float32
	FiberGrams            float32
	ProteinGrams          float32
	SaturatedFatGrams     float32
	ServingSize           string
	SodiumMilligrams      float32
	SugarGrams            float32
	TransFatGrams         float32
	UnsaturatedFatGrams   float32
}

// Diet is a diet restricted to certain foods or preparations.
type Diet int

const (
	UnknownDiet Diet = iota
	DiabeticDiet
	GlutenFreeDiet
	HalalDiet
	HinduDiet
	KosherDiet
	LowCalorieDiet
	LowFatDiet
	LowLactoseDiet
	LowSaltDiet
	VeganDiet
	VegetarianDiet
)
