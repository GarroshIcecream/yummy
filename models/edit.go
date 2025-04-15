package models

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	db "github.com/GarroshIcecream/yummy/db"
	recipes "github.com/GarroshIcecream/yummy/recipe"
	tea "github.com/charmbracelet/bubbletea"
	huh "github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

type EditModel struct {
	cookbook  *db.CookBook
	form      *huh.Form
	recipe    *recipes.RecipeRaw
	recipe_id uint
	isNew     bool
	err       error
}

type SaveMsg struct {
	recipe *recipes.RecipeRaw
	err    error
}

func NewEditModel(cookbook db.CookBook, recipe *recipes.RecipeRaw, recipe_id uint) *EditModel {
	// Get ingredients for the recipe
	var ingredients []recipes.Ingredient
	if recipe != nil {
		ingredients = recipe.Ingredients
	}

	// Generate fields for all ingredients
	ingredientFields := generateIngredientFields(ingredients)

	// Create all fields for the form
	allFields := []huh.Field{
		huh.NewInput().
			Key("name").
			Title("Recipe Name").
			Value(&[]string{""}[0]).
			Validate(func(value string) error {
				if value == "" {
					return fmt.Errorf("name is required")
				}
				if len(value) > 100 {
					return fmt.Errorf("name must be less than 100 characters")
				}
				return nil
			}),

		huh.NewInput().
			Key("description").
			Title("Description").
			Value(ternaryPtr(recipe != nil, &recipe.Description, &[]string{""}[0])).
			Validate(func(value string) error {
				if len(value) > 500 {
					return fmt.Errorf("description must be less than 500 characters")
				}
				return nil
			}),

		huh.NewInput().
			Key("author").
			Title("Author").
			Value(ternaryPtr(recipe != nil, &recipe.Author, &[]string{""}[0])).
			Validate(func(value string) error {
				if len(value) > 100 {
					return fmt.Errorf("author must be less than 100 characters")
				}
				return nil
			}),

		huh.NewInput().
			Key("prepTime").
			Title("Prep Time (minutes)").
			Value(ternaryPtr(recipe != nil, &[]string{fmt.Sprintf("%d", int(recipe.PrepTime.Minutes()))}[0], &[]string{""}[0])).
			Validate(func(value string) error {
				if value == "" {
					return fmt.Errorf("prep time is required")
				}
				_, err := strconv.Atoi(value)
				if err != nil {
					return fmt.Errorf("prep time must be a number")
				}
				return nil
			}),

		huh.NewInput().
			Key("cookTime").
			Title("Cook Time (minutes)").
			Value(ternaryPtr(recipe != nil, &[]string{fmt.Sprintf("%d", int(recipe.CookTime.Minutes()))}[0], &[]string{""}[0])).
			Validate(func(value string) error {
				if value == "" {
					return fmt.Errorf("cook time is required")
				}
				_, err := strconv.Atoi(value)
				if err != nil {
					return fmt.Errorf("cook time must be a number")
				}
				return nil
			}),

		huh.NewInput().
			Key("servings").
			Title("Servings").
			Value(ternaryPtr(recipe != nil, &recipe.Quantity, &[]string{""}[0])).
			Validate(func(value string) error {
				if value == "" {
					return fmt.Errorf("servings is required")
				}
				if len(value) > 10 {
					return fmt.Errorf("servings must be less than 10 characters")
				}
				return nil
			}),

		huh.NewInput().
			Key("url").
			Title("URL").
			Value(ternaryPtr(recipe != nil, &recipe.URL, &[]string{""}[0])).
			Validate(func(value string) error {
				if len(value) > 200 {
					return fmt.Errorf("URL must be less than 200 characters")
				}
				return nil
			}),

		huh.NewInput().
			Key("categories").
			Title("Categories (comma separated)").
			Value(ternaryPtr(recipe != nil, &[]string{fmt.Sprintf("%v", recipe.Categories)}[0], &[]string{""}[0])).
			Validate(func(value string) error {
				if len(value) > 200 {
					return fmt.Errorf("categories must be less than 200 characters")
				}
				return nil
			}),
	}

	// Add all ingredient fields to the allFields slice
	allFields = append(allFields, ingredientFields...)

	// Create the form with all fields in a single group
	form := huh.NewForm(huh.NewGroup(allFields...))

	return &EditModel{
		cookbook:  &cookbook,
		form:      form,
		recipe:    recipe,
		recipe_id: recipe_id,
		isNew:     recipe == nil,
	}
}

// Helper function to replace ternary operator for pointers
func ternaryPtr[T any](condition bool, trueVal, falseVal *T) *T {
	if condition {
		return trueVal
	}
	return falseVal
}

func (m *EditModel) Init() tea.Cmd {
	if !m.isNew {
		return tea.Cmd(m.loadRecipe)
	}
	return m.form.Init()
}

func (m *EditModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return NewListModel(*m.cookbook, nil), nil
		case tea.KeyEnter:
			if m.form.State == huh.StateCompleted {
				return m, tea.Cmd(m.saveRecipe)
			}
		}

	case SaveMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		return NewListModel(*m.cookbook, nil), nil

	case LoadRecipeMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.recipe = msg.recipe
		// Create a new form with the recipe data
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Key("name").
					Title("Recipe Name").
					Value(&m.recipe.Name).
					Validate(func(value string) error {
						if value == "" {
							return fmt.Errorf("name is required")
						}
						if len(value) > 100 {
							return fmt.Errorf("name must be less than 100 characters")
						}
						return nil
					}),

				huh.NewInput().
					Key("description").
					Title("Description").
					Value(&m.recipe.Description).
					Validate(func(value string) error {
						if len(value) > 500 {
							return fmt.Errorf("description must be less than 500 characters")
						}
						return nil
					}),

				huh.NewInput().
					Key("author").
					Title("Author").
					Value(&m.recipe.Author).
					Validate(func(value string) error {
						if len(value) > 100 {
							return fmt.Errorf("author must be less than 100 characters")
						}
						return nil
					}),

				huh.NewInput().
					Key("prepTime").
					Title("Prep Time (minutes)").
					Value(&[]string{fmt.Sprintf("%d", int(m.recipe.PrepTime.Minutes()))}[0]).
					Validate(func(value string) error {
						if value == "" {
							return fmt.Errorf("prep time is required")
						}
						_, err := strconv.Atoi(value)
						if err != nil {
							return fmt.Errorf("prep time must be a number")
						}
						return nil
					}),

				huh.NewInput().
					Key("cookTime").
					Title("Cook Time (minutes)").
					Value(&[]string{fmt.Sprintf("%d", int(m.recipe.CookTime.Minutes()))}[0]).
					Validate(func(value string) error {
						if value == "" {
							return fmt.Errorf("cook time is required")
						}
						_, err := strconv.Atoi(value)
						if err != nil {
							return fmt.Errorf("cook time must be a number")
						}
						return nil
					}),

				huh.NewInput().
					Key("servings").
					Title("Servings").
					Value(&m.recipe.Quantity).
					Validate(func(value string) error {
						if value == "" {
							return fmt.Errorf("servings is required")
						}
						if len(value) > 10 {
							return fmt.Errorf("servings must be less than 10 characters")
						}
						return nil
					}),

				huh.NewInput().
					Key("url").
					Title("URL").
					Value(&m.recipe.URL).
					Validate(func(value string) error {
						if len(value) > 200 {
							return fmt.Errorf("URL must be less than 200 characters")
						}
						return nil
					}),

				huh.NewInput().
					Key("categories").
					Title("Categories (comma separated)").
					Value(&[]string{fmt.Sprintf("%v", m.recipe.Categories)}[0]).
					Validate(func(value string) error {
						if len(value) > 200 {
							return fmt.Errorf("categories must be less than 200 characters")
						}
						return nil
					}),

				huh.NewNote().
					Title("Ingredients").
					Description("Add or edit ingredients"),

				huh.NewInput().
					Key("ingredient_amount").
					Title("Amount").
					Value(&[]string{""}[0]).
					Validate(func(value string) error {
						if len(value) > 10 {
							return fmt.Errorf("amount must be less than 10 characters")
						}
						return nil
					}),

				huh.NewInput().
					Key("ingredient_unit").
					Title("Unit").
					Value(&[]string{""}[0]).
					Validate(func(value string) error {
						if len(value) > 20 {
							return fmt.Errorf("unit must be less than 20 characters")
						}
						return nil
					}),

				huh.NewInput().
					Key("ingredient_name").
					Title("Name").
					Value(&[]string{""}[0]).
					Validate(func(value string) error {
						if value == "" {
							return fmt.Errorf("name is required")
						}
						if len(value) > 100 {
							return fmt.Errorf("name must be less than 100 characters")
						}
						return nil
					}),

				huh.NewInput().
					Key("ingredient_details").
					Title("Details (optional)").
					Value(&[]string{""}[0]).
					Validate(func(value string) error {
						if len(value) > 200 {
							return fmt.Errorf("details must be less than 200 characters")
						}
						return nil
					}),
			),
		)
		m.form = form
		return m, m.form.Init()
	}

	// Handle form updates
	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
	}

	return m, cmd
}

func (m *EditModel) saveRecipe() tea.Msg {
	// Parse prep time
	prepTime, err := strconv.Atoi(m.form.GetString("prepTime"))
	if err != nil {
		return SaveMsg{err: fmt.Errorf("invalid prep time: %v", err)}
	}

	// Parse cook time
	cookTime, err := strconv.Atoi(m.form.GetString("cookTime"))
	if err != nil {
		return SaveMsg{err: fmt.Errorf("invalid cook time: %v", err)}
	}

	// Collect all ingredients
	var ingredients []recipes.Ingredient
	i := 0
	for {
		// Try to get the ingredient fields with the current index
		prefix := fmt.Sprintf("ingredient_%d_", i)
		amount := m.form.GetString(prefix + "amount")
		unit := m.form.GetString(prefix + "unit")
		name := m.form.GetString(prefix + "name")
		details := m.form.GetString(prefix + "details")

		// If we can't find the name field (required), we've reached the end
		if name == "" {
			break
		}

		// Add the ingredient to our list
		ingredients = append(ingredients, recipes.Ingredient{
			Amount:  amount,
			Unit:    unit,
			Name:    name,
			Details: details,
		})

		i++
	}

	// Create recipe
	recipe := &recipes.RecipeRaw{
		Name:        m.form.GetString("name"),
		Description: m.form.GetString("description"),
		Author:      m.form.GetString("author"),
		PrepTime:    time.Duration(prepTime) * time.Minute,
		CookTime:    time.Duration(cookTime) * time.Minute,
		TotalTime:   time.Duration(prepTime+cookTime) * time.Minute,
		Quantity:    m.form.GetString("servings"),
		URL:         m.form.GetString("url"),
		Categories:  splitCategories(m.form.GetString("categories")),
		Ingredients: ingredients,
	}

	// Save to database
	var saveErr error
	if m.isNew {
		_, saveErr = m.cookbook.SaveScrapedRecipe(recipe)
	} else {
		// TODO: Implement UpdateRecipe in the database package
		saveErr = fmt.Errorf("updating recipes not yet implemented")
	}

	return SaveMsg{recipe: recipe, err: saveErr}
}

func (m *EditModel) loadRecipe() tea.Msg {
	recipe, err := m.cookbook.GetFullRecipe(m.recipe_id)
	if err != nil {
		return LoadRecipeMsg{err: fmt.Errorf("failed to load recipe: %v", err)}
	}
	return LoadRecipeMsg{recipe: recipe}
}

func (m *EditModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v", m.err)
	}

	var s strings.Builder

	title := "📝 Edit Recipe"
	if m.isNew {
		title = "📝 New Recipe"
	}
	s.WriteString(title + "\n\n")

	s.WriteString(m.form.View())

	return lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#FF6B6B")).
		Padding(1).
		Render(s.String())
}

func splitCategories(categories string) []string {
	if categories == "" {
		return nil
	}
	parts := strings.Split(categories, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

// generateIngredientFields creates form fields for a list of ingredients
func generateIngredientFields(ingredients []recipes.Ingredient) []huh.Field {
	fields := make([]huh.Field, 0)

	// Add a note for the ingredients section
	fields = append(fields, huh.NewNote().
		Title("Ingredients").
		Description("Add or edit ingredients"))

	// If no ingredients, add a default empty ingredient
	if len(ingredients) == 0 {
		ingredients = []recipes.Ingredient{{}}
	}

	// Add fields for each ingredient
	for i, ingredient := range ingredients {
		// Create a unique prefix for each ingredient to avoid key conflicts
		prefix := fmt.Sprintf("ingredient_%d_", i)

		// Add amount field
		fields = append(fields, huh.NewInput().
			Key(prefix+"amount").
			Title(fmt.Sprintf("Amount (%d)", i+1)).
			Value(&[]string{ingredient.Amount}[0]).
			Validate(func(value string) error {
				if len(value) > 10 {
					return fmt.Errorf("amount must be less than 10 characters")
				}
				return nil
			}))

		// Add unit field
		fields = append(fields, huh.NewInput().
			Key(prefix+"unit").
			Title(fmt.Sprintf("Unit (%d)", i+1)).
			Value(&[]string{ingredient.Unit}[0]).
			Validate(func(value string) error {
				if len(value) > 20 {
					return fmt.Errorf("unit must be less than 20 characters")
				}
				return nil
			}))

		// Add name field
		fields = append(fields, huh.NewInput().
			Key(prefix+"name").
			Title(fmt.Sprintf("Name (%d)", i+1)).
			Value(&[]string{ingredient.Name}[0]).
			Validate(func(value string) error {
				if value == "" {
					return fmt.Errorf("name is required")
				}
				if len(value) > 100 {
					return fmt.Errorf("name must be less than 100 characters")
				}
				return nil
			}))

		// Add details field
		fields = append(fields, huh.NewInput().
			Key(prefix+"details").
			Title(fmt.Sprintf("Details (optional) (%d)", i+1)).
			Value(&[]string{ingredient.Details}[0]).
			Validate(func(value string) error {
				if len(value) > 200 {
					return fmt.Errorf("details must be less than 200 characters")
				}
				return nil
			}))
	}

	return fields
}

var (
	focusedButton = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF6B6B")).
			Render("[ Save ]")

	blurredButton = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A0A0A0")).
			Render("[ Save ]")
)
