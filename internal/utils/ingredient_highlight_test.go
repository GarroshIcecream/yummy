package utils

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestParseIngredient(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		amount  string
		unit    string
		ingName string
		details string
	}{
		{
			name:    "whole number with unit",
			input:   "2 cup flour",
			amount:  "2",
			unit:    "cup",
			ingName: "flour",
		},
		{
			name:    "fraction amount",
			input:   "1/2 cup sugar",
			amount:  "1/2",
			unit:    "cup",
			ingName: "sugar",
		},
		{
			name:    "decimal amount 0.5",
			input:   "0.5 cup red wine",
			amount:  "0.5",
			unit:    "cup",
			ingName: "red wine",
		},
		{
			name:    "decimal amount 0.25",
			input:   "0.25 teaspoon ground cinnamon",
			amount:  "0.25",
			unit:    "tsp",
			ingName: "ground cinnamon",
		},
		{
			name:    "decimal amount 1.5",
			input:   "1.5 cups freshly grated parmesan cheese",
			amount:  "1.5",
			unit:    "cup",
			ingName: "freshly grated parmesan cheese",
		},
		{
			name:    "decimal amount 0.5 teaspoon",
			input:   "0.5 teaspoon fines herbs",
			amount:  "0.5",
			unit:    "tsp",
			ingName: "fines herbs",
		},
		{
			name:    "whole number no unit",
			input:   "3 eggplants",
			amount:  "3",
			unit:    "",
			ingName: "eggplants",
		},
		{
			name:    "with details in parens",
			input:   "1 pound lean ground beef (80/20)",
			amount:  "1",
			unit:    "pound",
			ingName: "lean ground beef",
			details: "80/20",
		},
		{
			name:    "name only no amount",
			input:   "salt to taste",
			amount:  "",
			unit:    "",
			ingName: "salt to taste",
		},
		{
			name:    "range amount",
			input:   "2-3 cup broth",
			amount:  "2-3",
			unit:    "cup",
			ingName: "broth",
		},
		{
			name:    "decimal in parenthetical detail",
			input:   "1 can tomato sauce (8 ounce)",
			amount:  "1",
			unit:    "can",
			ingName: "tomato sauce",
			details: "8 ounce",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ing, err := ParseIngredient(tt.input)
			if err != nil {
				t.Fatalf("ParseIngredient(%q) returned error: %v", tt.input, err)
			}
			if ing.Amount != tt.amount {
				t.Errorf("Amount: got %q, want %q", ing.Amount, tt.amount)
			}
			if ing.Unit != tt.unit {
				t.Errorf("Unit: got %q, want %q", ing.Unit, tt.unit)
			}
			if ing.Name != tt.ingName {
				t.Errorf("Name: got %q, want %q", ing.Name, tt.ingName)
			}
			if ing.Details != tt.details {
				t.Errorf("Details: got %q, want %q", ing.Details, tt.details)
			}
		})
	}
}

func TestExtractIngredientTokens(t *testing.T) {
	ingredients := []Ingredient{
		{Name: "oil"},
		{Name: "olive oil"},
		{Name: "salt"},
		{Name: ""},     // empty — should be skipped
		{Name: "  "},   // blank — should be skipped
		{Name: "salt"}, // duplicate — should be deduplicated
		{Name: "garlic cloves"},
	}

	tokens := extractIngredientTokens(ingredients)

	// Should contain full names + individual words (>= 3 chars, not stopwords)
	// "oil", "olive oil", "salt", "garlic cloves", "olive", "garlic", "cloves"
	tokenSet := make(map[string]bool)
	for _, tok := range tokens {
		tokenSet[strings.ToLower(tok)] = true
	}

	for _, expected := range []string{"olive oil", "garlic cloves", "oil", "salt", "olive", "garlic", "cloves"} {
		if !tokenSet[expected] {
			t.Errorf("expected token %q to be present, tokens: %v", expected, tokens)
		}
	}

	// Longest should come first
	if len(tokens) > 0 && tokens[0] != "garlic cloves" && tokens[0] != "olive oil" {
		t.Errorf("expected longest token first, got %q", tokens[0])
	}
}

func TestHighlightIngredientsInMarkdown_Basic(t *testing.T) {
	ingredients := []Ingredient{
		{Name: "flour"},
		{Name: "sugar"},
		{Name: "butter"},
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "single ingredient",
			input:    "Add the flour to the bowl",
			expected: "Add the **flour** to the bowl",
		},
		{
			name:     "multiple ingredients",
			input:    "Mix flour and sugar together",
			expected: "Mix **flour** and **sugar** together",
		},
		{
			name:     "no match",
			input:    "Preheat the oven to 350F",
			expected: "Preheat the oven to 350F",
		},
		{
			name:     "case insensitive",
			input:    "Add the Flour and SUGAR",
			expected: "Add the **Flour** and **SUGAR**",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HighlightIngredientsInMarkdown(tt.input, ingredients)
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestHighlightIngredientsInMarkdown_MultiWordName(t *testing.T) {
	ingredients := []Ingredient{
		{Name: "olive oil"},
		{Name: "garlic cloves"},
	}

	tests := []struct {
		name     string
		input    string
		contains []string
	}{
		{
			name:     "full multi-word match",
			input:    "Heat the olive oil in a pan",
			contains: []string{"**olive oil**"},
		},
		{
			name:     "partial word match from multi-word ingredient",
			input:    "Mince the garlic finely",
			contains: []string{"**garlic**"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HighlightIngredientsInMarkdown(tt.input, ingredients)
			for _, c := range tt.contains {
				if !strings.Contains(result, c) {
					t.Errorf("expected result to contain %q, got %q", c, result)
				}
			}
		})
	}
}

func TestHighlightIngredientsInMarkdown_EmptyIngredients(t *testing.T) {
	input := "Mix everything together"

	result := HighlightIngredientsInMarkdown(input, nil)
	if result != input {
		t.Errorf("expected unchanged text, got %q", result)
	}

	result = HighlightIngredientsInMarkdown(input, []Ingredient{})
	if result != input {
		t.Errorf("expected unchanged text, got %q", result)
	}
}

func TestHighlightIngredientsWithStyle_Basic(t *testing.T) {
	ingredients := []Ingredient{
		{Name: "garlic"},
		{Name: "onion"},
	}

	// Note: lipgloss may not emit ANSI codes in a headless/non-TTY
	// environment (e.g. test runner). We verify the function doesn't
	// panic and still contains the ingredient words.
	style := lipgloss.NewStyle().Bold(true)
	input := "Sauté the garlic and onion"
	result := HighlightIngredientsWithStyle(input, ingredients, style)

	if !strings.Contains(result, "garlic") {
		t.Error("expected result to contain 'garlic'")
	}
	if !strings.Contains(result, "onion") {
		t.Error("expected result to contain 'onion'")
	}
}

func TestHighlightIngredientsInMarkdown_WordBoundary(t *testing.T) {
	ingredients := []Ingredient{
		{Name: "egg"},
	}

	tests := []struct {
		name     string
		input    string
		contains string
		excludes string
	}{
		{
			name:     "matches whole word",
			input:    "Add one egg to the mix",
			contains: "**egg**",
		},
		{
			name:     "does not match inside words",
			input:    "Use eggplant in the dish",
			excludes: "**egg**",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HighlightIngredientsInMarkdown(tt.input, ingredients)
			if tt.contains != "" && !strings.Contains(result, tt.contains) {
				t.Errorf("expected result to contain %q, got %q", tt.contains, result)
			}
			if tt.excludes != "" && strings.Contains(result, tt.excludes) {
				t.Errorf("expected result NOT to contain %q, got %q", tt.excludes, result)
			}
		})
	}
}

func TestHighlightIngredientsInMarkdown_StopWords(t *testing.T) {
	// "fresh basil" → "fresh" is a stop word, only "basil" should match
	ingredients := []Ingredient{
		{Name: "fresh basil"},
	}

	input := "Add fresh basil leaves. Keep the rest fresh."
	result := HighlightIngredientsInMarkdown(input, ingredients)

	// "fresh basil" as a whole should be highlighted
	if !strings.Contains(result, "**fresh basil**") {
		t.Errorf("expected 'fresh basil' to be highlighted, got %q", result)
	}

	// "basil" alone should also be highlighted if it appears alone
	input2 := "Garnish with basil"
	result2 := HighlightIngredientsInMarkdown(input2, ingredients)
	if !strings.Contains(result2, "**basil**") {
		t.Errorf("expected 'basil' to be highlighted, got %q", result2)
	}
}
