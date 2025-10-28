package edit

import (
	"log/slog"
	"strings"
	"time"

	"github.com/GarroshIcecream/yummy/yummy/config"
	consts "github.com/GarroshIcecream/yummy/yummy/consts"
	db "github.com/GarroshIcecream/yummy/yummy/db"
	messages "github.com/GarroshIcecream/yummy/yummy/models/msg"
	recipes "github.com/GarroshIcecream/yummy/yummy/recipe"
	themes "github.com/GarroshIcecream/yummy/yummy/themes"
	utils "github.com/GarroshIcecream/yummy/yummy/utils"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
)

type EditState int

const (
	EditStateMainForm EditState = iota
	EditStateIngredients
	EditStateInstructions
	EditStateCompleted
)

type EditModel struct {
	// Configuration
	cookbook   *db.CookBook
	modelState consts.ModelState
	theme      *themes.Theme
	keyMap     config.KeyMap

	// Recipe
	recipeID *uint
	isNew    bool
	state    EditState
	width    int
	height   int

	// Form fields
	name        string
	description string
	author      string
	prepTime    string
	cookTime    string
	servings    string
	url         string
	categories  []string

	// Ingredients and instructions
	ingredients  []utils.Ingredient
	instructions []string

	// Forms
	mainForm        *huh.Form
	ingredientForm  *huh.Form
	instructionForm *huh.Form

	// Current ingredient being edited
	editingIngredientIndex int

	// Current instruction being edited
	editingInstructionIndex int

	// Navigation
	showHelp bool
}

func New(cookbook *db.CookBook, keymaps config.KeyMap, theme *themes.Theme, recipeID uint) *EditModel {
	model := &EditModel{
		cookbook:                cookbook,
		keyMap:                  keymaps,
		recipeID:                &recipeID,
		isNew:                   recipeID == 0,
		state:                   EditStateMainForm,
		showHelp:                false,
		modelState:              consts.ModelStateLoaded,
		theme:                   theme,
		ingredients:             []utils.Ingredient{},
		instructions:            []string{},
		editingIngredientIndex:  -1,
		editingInstructionIndex: -1,
	}

	if !model.isNew {
		recipe, err := model.FetchRecipe(*model.recipeID)
		if err != nil {
			slog.Error("Failed to fetch recipe: %s", "error", err)
			return nil
		}
		model.loadRecipe(recipe)
	}

	model.setupForms()
	return model
}

func (m *EditModel) Init() tea.Cmd {
	return m.mainForm.Init()
}

func (m *EditModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case messages.SaveMsg:
		cmds = append(cmds, messages.SendSessionStateMsg(consts.SessionStateDetail))
		if m.recipeID != nil {
			cmds = append(cmds, messages.SendRecipeSelectedMsg(*m.recipeID))
		}
		return m, tea.Batch(cmds...)

	case messages.EditRecipeMsg:
		// Load the recipe data and reinitialize the form
		recipe, err := m.FetchRecipe(msg.RecipeID)
		if err != nil {
			slog.Error("Failed to fetch recipe for editing: %s", "error", err)
			return m, nil
		}
		m.loadRecipe(recipe)
		m.setupForms()
		// Re-initialize the form to ensure it's properly set up
		err = m.mainForm.Run()
		if err != nil {
			slog.Error("Failed to run form: %s", "error", err)
			return m, nil
		}

	case messages.LoadRecipeMsg:
		m.loadRecipe(msg.Recipe)
		m.setupForms()
	}

	form, cmd := m.mainForm.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.mainForm = f

		// Check if form is completed and handle submission
		if m.mainForm.State == huh.StateCompleted {
			// Save the recipe
			_, err := m.saveRecipe()
			if err != nil {
				slog.Error("Failed to save recipe: %s", "error", err)
				// TODO: Show error message to user
			} else {
				// Send save message to return to detail view
				cmds = append(cmds, func() tea.Msg {
					return messages.SaveMsg{}
				})
			}
		}
	}
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *EditModel) View() string {
	var content strings.Builder

	title := "📝 Edit Recipe"
	if m.isNew {
		title = "📝 New Recipe"
	}
	content.WriteString(m.theme.DetailHeader.Render(title))
	content.WriteString("\n")

	return m.mainForm.View()
}

func (m *EditModel) FetchRecipe(recipeID uint) (*recipes.RecipeRaw, error) {
	recipe, err := m.cookbook.GetFullRecipe(recipeID)
	if err != nil {
		slog.Error("Failed to fetch recipe: %s", "error", err)
		return nil, err
	}
	return recipe, nil
}

func (m *EditModel) loadRecipe(recipe *recipes.RecipeRaw) {
	m.recipeID = &recipe.ID
	m.name = recipe.Name
	m.description = recipe.Description
	m.author = recipe.Author
	m.prepTime = recipe.PrepTime.String()
	m.cookTime = recipe.CookTime.String()
	m.servings = recipe.Quantity
	m.url = recipe.URL
	m.categories = recipe.Categories
	m.ingredients = recipe.Ingredients
	m.instructions = recipe.Instructions
}

func (m *EditModel) saveRecipe() (tea.Msg, error) {
	// Create recipe
	prepTime, err := time.ParseDuration(m.prepTime)
	if err != nil {
		return nil, err
	}
	cookTime, err := time.ParseDuration(m.cookTime)
	if err != nil {
		return nil, err
	}

	recipe := &recipes.RecipeRaw{
		Name:         m.name,
		Description:  m.description,
		Author:       m.author,
		PrepTime:     prepTime,
		CookTime:     cookTime,
		TotalTime:    prepTime + cookTime,
		Quantity:     m.servings,
		URL:          m.url,
		Categories:   m.categories,
		Ingredients:  m.ingredients,
		Instructions: m.instructions,
	}

	// Save to database
	if m.isNew {
		recipeID, err := m.cookbook.SaveScrapedRecipe(recipe)
		if err != nil {
			return nil, err
		}
		recipe.ID = recipeID
	} else {
		recipe.ID = *m.recipeID
		err := m.cookbook.UpdateRecipe(recipe)
		if err != nil {
			return nil, err
		}
	}

	return nil, nil
}

func (m *EditModel) setupForms() {
	all_categories, err := m.cookbook.GetAllCategories()
	if err != nil {
		slog.Error("Failed to get all categories: %s", "error", err)
	}

	categories_options := make([]huh.Option[string], len(all_categories))
	for i, category := range all_categories {
		categories_options[i] = huh.NewOption(category, category)
	}

	all_authors, err := m.cookbook.GetAllAuthors()
	if err != nil {
		slog.Error("Failed to get all authors: %s", "error", err)
	}

	// Main recipe form
	m.mainForm = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Key("name").
				Title("Recipe Name").
				Description("Enter the name of your recipe").
				Value(&m.name).
				Validate(utils.ValidateRequired),

			huh.NewText().
				Key("description").
				Title("Description").
				Description("Describe your recipe").
				Value(&m.description).
				Placeholder("A delicious recipe that..."),

			huh.NewInput().
				Key("author").
				Title("Author").
				Description("Who created this recipe?").
				Value(&m.author).
				Suggestions(all_authors),

			huh.NewInput().
				Key("prepTime").
				Title("Prep Time").
				Description("Preparation time in hours and minutes (e.g., '1h 30m', '2h', '30m')").
				Value(&m.prepTime).
				Validate(utils.ValidateDuration).
				Placeholder("10"),

			huh.NewInput().
				Key("cookTime").
				Title("Cook Time").
				Description("Cooking time in hours and minutes (e.g., '1h 30m', '2h', '30m')").
				Value(&m.cookTime).
				Validate(utils.ValidateDuration).
				Placeholder("10"),

			huh.NewInput().
				Key("servings").
				Title("Servings").
				Description("Number of servings (e.g., '4 servings', '2-3 people')").
				Value(&m.servings).
				Placeholder("4 servings"),

			huh.NewInput().
				Key("url").
				Title("Recipe URL").
				Description("Source URL (optional)").
				Value(&m.url).
				Placeholder("https://example.com/recipe").
				Validate(utils.ValidateURL),

			huh.NewMultiSelect[string]().
				Key("categories").
				Title("Categories").
				Description("Categories (e.g., 'dinner, italian, pasta')").
				Value(&m.categories).
				Options(categories_options...),
		),
	).
		WithTheme(huh.ThemeCharm()).
		WithWidth(80)

	// m.ingredientForm = huh.NewForm(
	// 	huh.NewGroup(
	// 		huh.NewInput().
	// 			Key("amount").
	// 			Title("Amount").
	// 			Description("Quantity (e.g., '2', '1/2', '1.5')").
	// 			Value(&m.currentIngredient.Amount).
	// 			Placeholder("2"),

	// 		huh.NewSelect[string]().
	// 			Key("unit").
	// 			Title("Unit").
	// 			Description("Measurement unit").
	// 			Options(
	// 				huh.NewOption("", ""),
	// 				huh.NewOption("cups", "cups"),
	// 				huh.NewOption("tablespoons", "tbsp"),
	// 				huh.NewOption("teaspoons", "tsp"),
	// 				huh.NewOption("pounds", "lbs"),
	// 				huh.NewOption("ounces", "oz"),
	// 				huh.NewOption("grams", "g"),
	// 				huh.NewOption("kilograms", "kg"),
	// 				huh.NewOption("milliliters", "ml"),
	// 				huh.NewOption("liters", "l"),
	// 				huh.NewOption("pieces", "pieces"),
	// 				huh.NewOption("cloves", "cloves"),
	// 				huh.NewOption("slices", "slices"),
	// 				huh.NewOption("pinch", "pinch"),
	// 				huh.NewOption("dash", "dash"),
	// 			).
	// 			Value(&m.currentIngredient.Unit),

	// 		huh.NewInput().
	// 			Key("name").
	// 			Title("Ingredient Name").
	// 			Description("Name of the ingredient").
	// 			Value(&m.currentIngredient.Name).
	// 			Validate(utils.ValidateRequired).
	// 			Placeholder("flour, salt, olive oil"),

	// 		huh.NewInput().
	// 			Key("details").
	// 			Title("Details (optional)").
	// 			Description("Additional details (e.g., 'finely chopped', 'room temperature')").
	// 			Value(&m.currentIngredient.Details).
	// 			Placeholder("finely chopped, room temperature"),
	// 	),
	// ).WithTheme(huh.ThemeCharm())

	// m.instructionForm = huh.NewForm(
	// 	huh.NewGroup(
	// 		huh.NewText().
	// 			Key("instruction").
	// 			Title("Cooking Step").
	// 			Description("Describe this cooking step").
	// 			Value(&m.currentInstruction).
	// 			Validate(utils.ValidateRequired),
	// 	),
	// ).WithTheme(huh.ThemeCharm())
}

func (m *EditModel) GetModelState() consts.ModelState {
	return m.modelState
}

func (m *EditModel) GetSessionState() consts.SessionState {
	return consts.SessionStateEdit
}

// GetSize returns the current width and height of the model
func (m *EditModel) GetSize() (width int, height int) {
	return m.width, m.height
}

// SetSize sets the width and height of the model
func (m *EditModel) SetSize(width, height int) {
	m.width = width
	m.height = height

	// Update form sizes to use full screen
	if m.mainForm != nil {
		m.mainForm = m.mainForm.WithWidth(width - 4) // Leave some margin
	}
	if m.ingredientForm != nil {
		m.ingredientForm = m.ingredientForm.WithWidth(width - 4)
	}
	if m.instructionForm != nil {
		m.instructionForm = m.instructionForm.WithWidth(width - 4)
	}
}
