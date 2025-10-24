package edit

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/GarroshIcecream/yummy/yummy/config"
	consts "github.com/GarroshIcecream/yummy/yummy/consts"
	db "github.com/GarroshIcecream/yummy/yummy/db"
	messages "github.com/GarroshIcecream/yummy/yummy/models/msg"
	recipes "github.com/GarroshIcecream/yummy/yummy/recipe"
	themes "github.com/GarroshIcecream/yummy/yummy/themes"
	"github.com/GarroshIcecream/yummy/yummy/utils"
	"github.com/charmbracelet/bubbles/key"
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
	cookbook *db.CookBook
	recipe   *recipes.RecipeRaw
	recipeID *uint
	isNew    bool
	err      error
	state    EditState
	width    int
	height   int
	keyMap   config.KeyMap

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
	ingredients  []recipes.Ingredient
	instructions []string

	// Forms
	mainForm        *huh.Form
	ingredientForm  *huh.Form
	instructionForm *huh.Form

	// Current ingredient being edited
	editingIngredientIndex int
	currentIngredient      recipes.Ingredient

	// Current instruction being edited
	editingInstructionIndex int
	currentInstruction      string

	// Navigation
	showHelp   bool
	modelState consts.ModelState
	theme      *themes.Theme
}

func New(cookbook *db.CookBook, keymaps config.KeyMap, theme *themes.Theme, recipe *recipes.RecipeRaw) *EditModel {
	model := &EditModel{
		cookbook:                cookbook,
		keyMap:                  keymaps,
		recipe:                  recipe,
		isNew:                   recipe == nil,
		state:                   EditStateMainForm,
		showHelp:                false,
		modelState:              consts.ModelStateLoaded,
		theme:                   theme,
		ingredients:             []recipes.Ingredient{},
		instructions:            []string{},
		editingIngredientIndex:  -1,
		editingInstructionIndex: -1,
	}

	if recipe != nil {
		model.recipeID = &recipe.ID
		model.name = recipe.Name
		model.description = recipe.Description
		model.author = recipe.Author
		model.prepTime = fmt.Sprintf("%d", int(recipe.PrepTime.Minutes()))
		model.cookTime = fmt.Sprintf("%d", int(recipe.CookTime.Minutes()))
		model.servings = recipe.Quantity
		model.url = recipe.URL
		model.categories = recipe.Categories
		model.ingredients = recipe.Ingredients
		model.instructions = recipe.Instructions
	}

	model.setupForms()
	return model
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
				Title("Prep Time (minutes)").
				Description("Preparation time in minutes").
				Value(&m.prepTime).
				Validate(utils.ValidateInteger).
				Placeholder("10"),

			huh.NewInput().
				Key("cookTime").
				Title("Cook Time (minutes)").
				Description("Cooking time in minutes").
				Value(&m.cookTime).
				Validate(utils.ValidateInteger).
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
	).WithTheme(huh.ThemeCharm()).WithWidth(80) // Default width

	// Ingredient editing form
	m.setupIngredientForm()

	// Instruction editing form
	m.setupInstructionForm()
}

func (m *EditModel) setupIngredientForm() {
	m.ingredientForm = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Key("amount").
				Title("Amount").
				Description("Quantity (e.g., '2', '1/2', '1.5')").
				Value(&m.currentIngredient.Amount).
				Placeholder("2"),

			huh.NewSelect[string]().
				Key("unit").
				Title("Unit").
				Description("Measurement unit").
				Options(
					huh.NewOption("", ""),
					huh.NewOption("cups", "cups"),
					huh.NewOption("tablespoons", "tbsp"),
					huh.NewOption("teaspoons", "tsp"),
					huh.NewOption("pounds", "lbs"),
					huh.NewOption("ounces", "oz"),
					huh.NewOption("grams", "g"),
					huh.NewOption("kilograms", "kg"),
					huh.NewOption("milliliters", "ml"),
					huh.NewOption("liters", "l"),
					huh.NewOption("pieces", "pieces"),
					huh.NewOption("cloves", "cloves"),
					huh.NewOption("slices", "slices"),
					huh.NewOption("pinch", "pinch"),
					huh.NewOption("dash", "dash"),
				).
				Value(&m.currentIngredient.Unit),

			huh.NewInput().
				Key("name").
				Title("Ingredient Name").
				Description("Name of the ingredient").
				Value(&m.currentIngredient.Name).
				Validate(utils.ValidateRequired).
				Placeholder("flour, salt, olive oil"),

			huh.NewInput().
				Key("details").
				Title("Details (optional)").
				Description("Additional details (e.g., 'finely chopped', 'room temperature')").
				Value(&m.currentIngredient.Details).
				Placeholder("finely chopped, room temperature"),
		),
	).WithTheme(huh.ThemeCharm()).WithWidth(80) // Default width
}

func (m *EditModel) setupInstructionForm() {
	m.instructionForm = huh.NewForm(
		huh.NewGroup(
			huh.NewText().
				Key("instruction").
				Title("Cooking Step").
				Description("Describe this cooking step").
				Value(&m.currentInstruction).
				Validate(utils.ValidateRequired),
		),
	).WithTheme(huh.ThemeCharm()).WithWidth(80) // Default width
}

func (m *EditModel) GetModelState() consts.ModelState {
	return m.modelState
}

func (m *EditModel) Init() tea.Cmd {
	if !m.isNew && m.recipeID != nil {
		return tea.Cmd(m.loadRecipe)
	}
	// Always initialize the main form
	return m.mainForm.Init()
}

func (m *EditModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keyMap.Back):
			if m.state == EditStateMainForm {
				// Go back to previous state
				cmds = append(cmds, messages.SendSessionStateMsg(consts.SessionStateDetail))
				if m.recipeID != nil {
					cmds = append(cmds, messages.SendRecipeSelectedMsg(*m.recipeID))
				}
			} else {
				m.state = EditStateMainForm
				return m, m.mainForm.Init()
			}
		case key.Matches(msg, m.keyMap.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keyMap.Help):
			m.showHelp = !m.showHelp
		}

		// Handle state-specific key messages
		switch m.state {
		case EditStateMainForm:
			// Handle main form updates and custom keys
			if updatedForm, cmd := m.mainForm.Update(msg); cmd != nil {
				m.mainForm = updatedForm.(*huh.Form)
				cmds = append(cmds, cmd)
			}
			cmds = append(cmds, m.handleMainFormKeys(msg))
		case EditStateIngredients:
			cmds = append(cmds, m.handleIngredientKeys(msg))
		case EditStateInstructions:
			cmds = append(cmds, m.handleInstructionKeys(msg))
		}

	case tea.WindowSizeMsg:
		m.SetSize(msg.Width, msg.Height)

	case messages.EditRecipeMsg:
		m.recipeID = &msg.RecipeID
		cmds = append(cmds, tea.Cmd(m.loadRecipe))

	case messages.LoadRecipeMsg:
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.recipe = msg.Recipe
		if m.recipe != nil {
			m.recipeID = &m.recipe.ID
			m.name = m.recipe.Name
			m.description = m.recipe.Description
			m.author = m.recipe.Author
			m.prepTime = fmt.Sprintf("%d", int(m.recipe.PrepTime.Minutes()))
			m.cookTime = fmt.Sprintf("%d", int(m.recipe.CookTime.Minutes()))
			m.servings = m.recipe.Quantity
			m.url = m.recipe.URL
			m.categories = m.recipe.Categories
			m.ingredients = m.recipe.Ingredients
			m.instructions = m.recipe.Instructions
			m.setupForms()
		}

	case messages.SaveMsg:
		if msg.Err != nil {
			m.err = msg.Err
		} else {
			// Recipe saved successfully, go back to detail view
			cmds = append(cmds, messages.SendSessionStateMsg(consts.SessionStateDetail))
			if m.recipeID != nil {
				cmds = append(cmds, messages.SendRecipeSelectedMsg(*m.recipeID))
			}
		}
	}

	return m, tea.Sequence(cmds...)
}

func (m *EditModel) handleMainFormKeys(msg tea.KeyMsg) tea.Cmd {
	switch {
	case key.Matches(msg, m.keyMap.EditIngredients):
		m.state = EditStateIngredients
		return nil
	case key.Matches(msg, m.keyMap.EditInstructions):
		m.state = EditStateInstructions
		return nil
	case key.Matches(msg, m.keyMap.Enter):
		// Run the form to completion
		return tea.Cmd(func() tea.Msg {
			err := m.mainForm.Run()
			if err != nil {
				return messages.SaveMsg{Err: fmt.Errorf("form validation failed: %v", err)}
			}
			// Form completed successfully, now save the recipe
			return m.saveRecipe()
		})
	}

	return nil
}

func (m *EditModel) handleIngredientKeys(msg tea.KeyMsg) tea.Cmd {
	switch {
	case key.Matches(msg, m.keyMap.EditAdd):
		// Add new ingredient
		m.currentIngredient = recipes.Ingredient{}
		m.editingIngredientIndex = len(m.ingredients)
		return tea.Cmd(func() tea.Msg {
			err := m.ingredientForm.Run()
			if err != nil {
				return messages.SaveMsg{Err: fmt.Errorf("ingredient form validation failed: %v", err)}
			}
			// Form completed successfully, save the ingredient
			if m.editingIngredientIndex >= len(m.ingredients) {
				m.ingredients = append(m.ingredients, m.currentIngredient)
			} else {
				m.ingredients[m.editingIngredientIndex] = m.currentIngredient
			}
			m.editingIngredientIndex = -1
			m.currentIngredient = recipes.Ingredient{}
			return nil
		})
	case key.Matches(msg, m.keyMap.EditEdit):
		if len(m.ingredients) > 0 {
			// Edit first ingredient (you could make this more sophisticated)
			m.currentIngredient = m.ingredients[0]
			m.editingIngredientIndex = 0
			return tea.Cmd(func() tea.Msg {
				err := m.ingredientForm.Run()
				if err != nil {
					return messages.SaveMsg{Err: fmt.Errorf("ingredient form validation failed: %v", err)}
				}
				// Form completed successfully, save the ingredient
				m.ingredients[m.editingIngredientIndex] = m.currentIngredient
				m.editingIngredientIndex = -1
				m.currentIngredient = recipes.Ingredient{}
				return nil
			})
		}
	case key.Matches(msg, m.keyMap.EditDelete):
		if len(m.ingredients) > 0 {
			// Delete first ingredient (you could make this more sophisticated)
			m.ingredients = m.ingredients[1:]
		}
	}

	return nil
}

func (m *EditModel) handleInstructionKeys(msg tea.KeyMsg) tea.Cmd {
	switch {
	case key.Matches(msg, m.keyMap.EditAdd):
		// Add new instruction
		m.currentInstruction = ""
		m.editingInstructionIndex = len(m.instructions)
		return tea.Cmd(func() tea.Msg {
			err := m.instructionForm.Run()
			if err != nil {
				return messages.SaveMsg{Err: fmt.Errorf("instruction form validation failed: %v", err)}
			}
			// Form completed successfully, save the instruction
			if m.editingInstructionIndex >= len(m.instructions) {
				m.instructions = append(m.instructions, m.currentInstruction)
			} else {
				m.instructions[m.editingInstructionIndex] = m.currentInstruction
			}
			m.editingInstructionIndex = -1
			m.currentInstruction = ""
			return nil
		})
	case key.Matches(msg, m.keyMap.EditEdit):
		if len(m.instructions) > 0 {
			// Edit first instruction (you could make this more sophisticated)
			m.currentInstruction = m.instructions[0]
			m.editingInstructionIndex = 0
			return tea.Cmd(func() tea.Msg {
				err := m.instructionForm.Run()
				if err != nil {
					return messages.SaveMsg{Err: fmt.Errorf("instruction form validation failed: %v", err)}
				}
				// Form completed successfully, save the instruction
				m.instructions[m.editingInstructionIndex] = m.currentInstruction
				m.editingInstructionIndex = -1
				m.currentInstruction = ""
				return nil
			})
		}
	case key.Matches(msg, m.keyMap.EditDelete):
		if len(m.instructions) > 0 {
			// Delete first instruction (you could make this more sophisticated)
			m.instructions = m.instructions[1:]
		}
	}

	return nil
}

func (m *EditModel) loadRecipe() tea.Msg {
	recipe, err := m.cookbook.GetFullRecipe(*m.recipeID)
	if err != nil {
		return messages.LoadRecipeMsg{Recipe: nil, Err: fmt.Errorf("failed to load recipe: %v", err)}
	}
	return messages.LoadRecipeMsg{Recipe: recipe}
}

func (m *EditModel) saveRecipe() tea.Msg {
	// Parse prep time
	var prepTime int
	if m.prepTime != "" {
		var err error
		prepTime, err = strconv.Atoi(m.prepTime)
		if err != nil {
			return messages.SaveMsg{Err: fmt.Errorf("invalid prep time: %v", err)}
		}
	}

	// Parse cook time
	var cookTime int
	if m.cookTime != "" {
		var err error
		cookTime, err = strconv.Atoi(m.cookTime)
		if err != nil {
			return messages.SaveMsg{Err: fmt.Errorf("invalid cook time: %v", err)}
		}
	}

	// Create recipe
	recipe := &recipes.RecipeRaw{
		Name:         m.name,
		Description:  m.description,
		Author:       m.author,
		PrepTime:     time.Duration(prepTime) * time.Minute,
		CookTime:     time.Duration(cookTime) * time.Minute,
		TotalTime:    time.Duration(prepTime+cookTime) * time.Minute,
		Quantity:     m.servings,
		URL:          m.url,
		Categories:   m.categories,
		Ingredients:  m.ingredients,
		Instructions: m.instructions,
	}

	// Set ID for updates
	if m.recipeID != nil {
		recipe.ID = *m.recipeID
	}

	// Save to database
	var saveErr error
	if m.isNew {
		recipeID, err := m.cookbook.SaveScrapedRecipe(recipe)
		if err != nil {
			saveErr = err
		} else {
			recipe.ID = recipeID
		}
	} else {
		saveErr = m.cookbook.UpdateRecipe(recipe)
	}

	return messages.SaveMsg{Recipe: recipe, Err: saveErr}
}

func (m *EditModel) View() string {
	if m.err != nil {
		return m.theme.Error.Render(fmt.Sprintf("❌ Error: %v", m.err))
	}

	var content strings.Builder

	title := "📝 Edit Recipe"
	if m.isNew {
		title = "📝 New Recipe"
	}
	content.WriteString(m.theme.DetailHeader.Render(title))
	content.WriteString("\n")

	switch m.state {
	case EditStateMainForm:
		content.WriteString(m.mainForm.View())
		// Add navigation hints
		content.WriteString("\n")
		content.WriteString(fmt.Sprintf("Press '%s' for ingredients, '%s' for instructions, %s to save, %s to cancel",
			m.keyMap.EditIngredients.Help().Key,
			m.keyMap.EditInstructions.Help().Key,
			m.keyMap.Enter.Help().Key,
			m.keyMap.Back.Help().Key))
	case EditStateIngredients:
		content.WriteString(m.renderIngredients())
	case EditStateInstructions:
		content.WriteString(m.renderInstructions())
	}

	// Add help
	if m.showHelp {
		content.WriteString("\n\n")
		content.WriteString(m.renderHelp())
	}

	return m.theme.Doc.Render(content.String())
}

func (m *EditModel) renderIngredients() string {
	var content strings.Builder

	content.WriteString("🥘 Ingredients\n\n")

	if len(m.ingredients) == 0 {
		content.WriteString("No ingredients added yet.\n\n")
	} else {
		for i, ingredient := range m.ingredients {
			content.WriteString(fmt.Sprintf("%d. ", i+1))
			if ingredient.Amount != "" && ingredient.Unit != "" {
				content.WriteString(fmt.Sprintf("%s %s ", ingredient.Amount, ingredient.Unit))
			} else if ingredient.Amount != "" {
				content.WriteString(fmt.Sprintf("%s ", ingredient.Amount))
			}
			content.WriteString(ingredient.Name)
			if ingredient.Details != "" {
				content.WriteString(fmt.Sprintf(" (%s)", ingredient.Details))
			}
			content.WriteString("\n")
		}
		content.WriteString("\n")
	}

	content.WriteString(fmt.Sprintf("Press '%s' to add, '%s' to edit first, '%s' to delete first, '%s' to go back",
		m.keyMap.EditAdd.Help().Key,
		m.keyMap.EditEdit.Help().Key,
		m.keyMap.EditDelete.Help().Key,
		m.keyMap.Back.Help().Key))

	return content.String()
}

func (m *EditModel) renderInstructions() string {
	var content strings.Builder

	content.WriteString("👩‍🍳 Instructions\n\n")

	if len(m.instructions) == 0 {
		content.WriteString("No instructions added yet.\n\n")
	} else {
		for i, instruction := range m.instructions {
			content.WriteString(fmt.Sprintf("%d. %s\n\n", i+1, instruction))
		}
	}

	content.WriteString(fmt.Sprintf("Press '%s' to add, '%s' to edit first, '%s' to delete first, '%s' to go back",
		m.keyMap.EditAdd.Help().Key,
		m.keyMap.EditEdit.Help().Key,
		m.keyMap.EditDelete.Help().Key,
		m.keyMap.Back.Help().Key))

	return content.String()
}

func (m *EditModel) renderHelp() string {
	help := []string{
		"Navigation:",
		fmt.Sprintf("  %s - Go back", m.keyMap.Back.Help().Key),
		fmt.Sprintf("  %s - Toggle help", m.keyMap.Help.Help().Key),
		"",
		"Main Form:",
		fmt.Sprintf("  %s - Edit ingredients", m.keyMap.EditIngredients.Help().Key),
		fmt.Sprintf("  %s - Edit instructions", m.keyMap.EditInstructions.Help().Key),
		fmt.Sprintf("  %s - Save recipe", m.keyMap.Enter.Help().Key),
		"",
		"Ingredients/Instructions:",
		fmt.Sprintf("  %s - Add new item", m.keyMap.EditAdd.Help().Key),
		fmt.Sprintf("  %s - Edit first item", m.keyMap.EditEdit.Help().Key),
		fmt.Sprintf("  %s - Delete first item", m.keyMap.EditDelete.Help().Key),
		fmt.Sprintf("  %s - Go back to main form", m.keyMap.Back.Help().Key),
	}

	return m.theme.Help.Render(strings.Join(help, "\n"))
}

func (m *EditModel) GetSessionState() consts.SessionState {
	return consts.SessionStateEdit
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

// GetSize returns the current width and height of the model
func (m *EditModel) GetSize() (width, height int) {
	return m.width, m.height
}
