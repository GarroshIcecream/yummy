package edit

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/GarroshIcecream/yummy/yummy/config"
	db "github.com/GarroshIcecream/yummy/yummy/db"
	recipes "github.com/GarroshIcecream/yummy/yummy/recipe"
	styles "github.com/GarroshIcecream/yummy/yummy/tui/styles"
	ui "github.com/GarroshIcecream/yummy/yummy/ui"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type EditState int

const (
	EditStateForm EditState = iota
	EditStateIngredients
	EditStateInstructions
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
	categories  string

	// Ingredients management
	ingredients    []recipes.Ingredient
	ingredientList list.Model

	// Instructions management
	instructions    []string
	instructionList list.Model

	// Navigation
	activeField int
	showHelp    bool
	modelState  ui.ModelState
}

func New(cookbook *db.CookBook, keymaps config.KeyMap, recipe *recipes.RecipeRaw) *EditModel {
	model := &EditModel{
		cookbook:     cookbook,
		keyMap:       keymaps,
		recipe:       recipe,
		isNew:        recipe == nil,
		state:        EditStateForm,
		activeField:  0,
		showHelp:     false,
		modelState:   ui.ModelStateLoaded,
		ingredients:  []recipes.Ingredient{},
		instructions: []string{},
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
		model.categories = strings.Join(recipe.Categories, ", ")
		model.ingredients = recipe.Ingredients
		model.instructions = recipe.Instructions
	} else {
		// Default values for new recipe
		model.ingredients = []recipes.Ingredient{{}}
		model.instructions = []string{""}
	}

	model.setupLists()
	return model
}

func (m *EditModel) GetModelState() ui.ModelState {
	return m.modelState
}

func (m *EditModel) setupLists() {
	// Setup ingredients list
	var ingredientItems []list.Item
	for i, ingredient := range m.ingredients {
		ingredientItems = append(ingredientItems, ingredientItem{
			index:      i,
			ingredient: ingredient,
		})
	}

	ingredientDelegate := list.NewDefaultDelegate()
	ingredientDelegate.Styles = styles.GetDelegateStyles()
	ingredientDelegate.ShowDescription = true
	ingredientDelegate.ShortHelpFunc = func() []key.Binding {
		return []key.Binding{
			m.keyMap.Enter,
			m.keyMap.Delete,
			m.keyMap.Add,
			m.keyMap.Back,
		}
	}

	m.ingredientList = list.New(ingredientItems, ingredientDelegate, 0, 0)
	m.ingredientList.Styles = styles.GetListStyles()
	m.ingredientList.Title = "Ingredients"
	m.ingredientList.SetStatusBarItemName("ingredient", "ingredients")

	// Setup instructions list
	var instructionItems []list.Item
	for i, instruction := range m.instructions {
		instructionItems = append(instructionItems, instructionItem{
			index:       i,
			instruction: instruction,
		})
	}

	instructionDelegate := list.NewDefaultDelegate()
	instructionDelegate.Styles = styles.GetDelegateStyles()
	instructionDelegate.ShowDescription = false
	instructionDelegate.ShortHelpFunc = func() []key.Binding {
		return []key.Binding{
			m.keyMap.Enter,
			m.keyMap.Delete,
			m.keyMap.Add,
			m.keyMap.Back,
		}
	}

	m.instructionList = list.New(instructionItems, instructionDelegate, 0, 0)
	m.instructionList.Styles = styles.GetListStyles()
	m.instructionList.Title = "Instructions"
	m.instructionList.SetStatusBarItemName("instruction", "instructions")
}

func (m *EditModel) Init() tea.Cmd {
	if !m.isNew && m.recipeID != nil {
		return tea.Cmd(m.loadRecipe)
	}
	return nil
}

func (m *EditModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keyMap.Back):
			if m.state != EditStateForm {
				m.state = EditStateForm
				return m, nil
			}
			// Go back to previous state
			cmds = append(cmds, ui.SendSessionStateMsg(ui.SessionStateDetail))
			if m.recipeID != nil {
				cmds = append(cmds, ui.SendRecipeSelectedMsg(*m.recipeID))
			}
		case key.Matches(msg, m.keyMap.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keyMap.Help):
			m.showHelp = !m.showHelp
		case key.Matches(msg, m.keyMap.Enter):
			if m.state == EditStateForm {
				cmds = append(cmds, tea.Cmd(m.saveRecipe))
			}
		}

		// Handle state-specific key messages
		switch m.state {
		case EditStateForm:
			cmds = append(cmds, m.handleFormKeys(msg))
		case EditStateIngredients:
			cmds = append(cmds, m.handleIngredientKeys(msg))
		case EditStateInstructions:
			cmds = append(cmds, m.handleInstructionKeys(msg))
		}

	case tea.WindowSizeMsg:
		m.SetSize(msg.Width, msg.Height)

	case ui.EditRecipeMsg:
		m.recipeID = &msg.RecipeID
		cmds = append(cmds, tea.Cmd(m.loadRecipe))

	case ui.LoadRecipeMsg:
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
			m.categories = strings.Join(m.recipe.Categories, ", ")
			m.ingredients = m.recipe.Ingredients
			m.instructions = m.recipe.Instructions
			m.setupLists()
		}

	case ui.SaveMsg:
		if msg.Err != nil {
			m.err = msg.Err
		} else {
			// Recipe saved successfully, go back to detail view
			cmds = append(cmds, ui.SendSessionStateMsg(ui.SessionStateDetail))
			if m.recipeID != nil {
				cmds = append(cmds, ui.SendRecipeSelectedMsg(*m.recipeID))
			}
		}
	}

	return m, tea.Sequence(cmds...)
}

func (m *EditModel) handleFormKeys(msg tea.KeyMsg) tea.Cmd {
	switch {
	case key.Matches(msg, m.keyMap.CursorUp):
		if m.activeField > 0 {
			m.activeField--
		}
	case key.Matches(msg, m.keyMap.CursorDown):
		if m.activeField < 7 { // 8 fields total (0-7)
			m.activeField++
		}
	case key.Matches(msg, key.NewBinding(key.WithKeys("i"))):
		m.state = EditStateIngredients
		return nil
	case key.Matches(msg, key.NewBinding(key.WithKeys("s"))):
		m.state = EditStateInstructions
		return nil
	}
	return nil
}

func (m *EditModel) handleIngredientKeys(msg tea.KeyMsg) tea.Cmd {
	var cmd tea.Cmd
	m.ingredientList, cmd = m.ingredientList.Update(msg)

	switch {
	case key.Matches(msg, m.keyMap.Add):
		// Add new ingredient
		m.ingredients = append(m.ingredients, recipes.Ingredient{})
		m.setupLists()
	case key.Matches(msg, m.keyMap.Delete):
		// Delete selected ingredient
		if selected := m.ingredientList.SelectedItem(); selected != nil {
			if item, ok := selected.(ingredientItem); ok {
				if len(m.ingredients) > 1 {
					m.ingredients = append(m.ingredients[:item.index], m.ingredients[item.index+1:]...)
					m.setupLists()
				}
			}
		}
	case key.Matches(msg, m.keyMap.Enter):
		// Edit selected ingredient
		if selected := m.ingredientList.SelectedItem(); selected != nil {
			if _, ok := selected.(ingredientItem); ok {
				// TODO: Implement ingredient editing dialog
				// For now, just return to form
				m.state = EditStateForm
			}
		}
	}

	return cmd
}

func (m *EditModel) handleInstructionKeys(msg tea.KeyMsg) tea.Cmd {
	var cmd tea.Cmd
	m.instructionList, cmd = m.instructionList.Update(msg)

	switch {
	case key.Matches(msg, m.keyMap.Add):
		// Add new instruction
		m.instructions = append(m.instructions, "")
		m.setupLists()
	case key.Matches(msg, m.keyMap.Delete):
		// Delete selected instruction
		if selected := m.instructionList.SelectedItem(); selected != nil {
			if item, ok := selected.(instructionItem); ok {
				if len(m.instructions) > 1 {
					m.instructions = append(m.instructions[:item.index], m.instructions[item.index+1:]...)
					m.setupLists()
				}
			}
		}
	case key.Matches(msg, m.keyMap.Enter):
		// Edit selected instruction
		if selected := m.instructionList.SelectedItem(); selected != nil {
			if _, ok := selected.(instructionItem); ok {
				// TODO: Implement instruction editing dialog
				// For now, just return to form
				m.state = EditStateForm
			}
		}
	}

	return cmd
}

func (m *EditModel) loadRecipe() tea.Msg {
	recipe, err := m.cookbook.GetFullRecipe(*m.recipeID)
	if err != nil {
		return ui.LoadRecipeMsg{Recipe: nil, Err: fmt.Errorf("failed to load recipe: %v", err)}
	}
	return ui.LoadRecipeMsg{Recipe: recipe}
}

func (m *EditModel) saveRecipe() tea.Msg {
	// Parse prep time
	prepTime, err := strconv.Atoi(m.prepTime)
	if err != nil {
		return ui.SaveMsg{Err: fmt.Errorf("invalid prep time: %v", err)}
	}

	// Parse cook time
	cookTime, err := strconv.Atoi(m.cookTime)
	if err != nil {
		return ui.SaveMsg{Err: fmt.Errorf("invalid cook time: %v", err)}
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
		Categories:   m.splitCategories(m.categories),
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

	return ui.SaveMsg{Recipe: recipe, Err: saveErr}
}

func (m *EditModel) splitCategories(categories string) []string {
	if categories == "" {
		return nil
	}
	parts := strings.Split(categories, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

func (m *EditModel) View() string {
	if m.err != nil {
		return styles.ErrorStyle.Render(fmt.Sprintf("❌ Error: %v", m.err))
	}

	var content strings.Builder

	title := "📝 Edit Recipe"
	if m.isNew {
		title = "📝 New Recipe"
	}
	content.WriteString(styles.DetailHeaderStyle.Render(title))
	content.WriteString("\n\n")

	switch m.state {
	case EditStateForm:
		content.WriteString(m.renderForm())
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

	return styles.DocStyle.Render(content.String())
}

func (m *EditModel) renderForm() string {
	var content strings.Builder

	fields := []struct {
		label       string
		value       *string
		placeholder string
	}{
		{"Name", &m.name, "Recipe name"},
		{"Description", &m.description, "Recipe description"},
		{"Author", &m.author, "Recipe author"},
		{"Prep Time (min)", &m.prepTime, "Preparation time in minutes"},
		{"Cook Time (min)", &m.cookTime, "Cooking time in minutes"},
		{"Servings", &m.servings, "Number of servings"},
		{"URL", &m.url, "Recipe URL"},
		{"Categories", &m.categories, "Comma-separated categories"},
	}

	for i, field := range fields {
		style := styles.InfoStyle
		if i == m.activeField {
			style = style.Bold(true).Foreground(lipgloss.Color("#FF6B6B"))
		}

		content.WriteString(style.Render(fmt.Sprintf("%s: %s", field.label, *field.value)))
		content.WriteString("\n")
	}

	content.WriteString("\n")
	content.WriteString(styles.HelpStyle.Render("Press 'i' for ingredients, 's' for instructions, Enter to save, Esc to cancel"))

	return content.String()
}

func (m *EditModel) renderIngredients() string {
	return m.ingredientList.View()
}

func (m *EditModel) renderInstructions() string {
	return m.instructionList.View()
}

func (m *EditModel) renderHelp() string {
	help := []string{
		"Navigation:",
		"  ↑/k - Move up",
		"  ↓/j - Move down",
		"  Enter - Select/Edit",
		"  Esc/q - Go back",
		"  ?/h - Toggle help",
		"",
		"Form:",
		"  i - Edit ingredients",
		"  s - Edit instructions",
		"  Enter - Save recipe",
		"",
		"Ingredients/Instructions:",
		"  a/+ - Add new item",
		"  x - Delete selected item",
		"  Enter - Edit selected item",
	}

	return styles.HelpStyle.Render(strings.Join(help, "\n"))
}

// SetSize sets the width and height of the model
func (m *EditModel) SetSize(width, height int) {
	m.width = width
	m.height = height

	// Update list sizes
	h, v := styles.DocStyle.GetFrameSize()
	listWidth := width - h
	listHeight := height - v - 10 // Reserve space for header and help

	m.ingredientList.SetSize(listWidth, listHeight)
	m.instructionList.SetSize(listWidth, listHeight)
}

// GetSize returns the current width and height of the model
func (m *EditModel) GetSize() (width, height int) {
	return m.width, m.height
}

// List item types for ingredients and instructions
type ingredientItem struct {
	index      int
	ingredient recipes.Ingredient
}

func (i ingredientItem) FilterValue() string {
	return i.ingredient.Name
}

func (i ingredientItem) Title() string {
	return i.ingredient.Name
}

func (i ingredientItem) Description() string {
	return fmt.Sprintf("%s %s - %s", i.ingredient.Amount, i.ingredient.Unit, i.ingredient.Details)
}

type instructionItem struct {
	index       int
	instruction string
}

func (i instructionItem) FilterValue() string {
	return i.instruction
}

func (i instructionItem) Title() string {
	return fmt.Sprintf("Step %d", i.index+1)
}

func (i instructionItem) Description() string {
	return i.instruction
}
