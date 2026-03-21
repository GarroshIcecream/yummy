package edit

import (
	"fmt"
	"log/slog"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"github.com/GarroshIcecream/yummy/internal/config"
	db "github.com/GarroshIcecream/yummy/internal/db"
	common "github.com/GarroshIcecream/yummy/internal/models/common"
	messages "github.com/GarroshIcecream/yummy/internal/models/msg"
	themes "github.com/GarroshIcecream/yummy/internal/themes"
	utils "github.com/GarroshIcecream/yummy/internal/utils"
)

type EditModel struct {
	// Configuration
	cookbook   *db.CookBook
	modelState common.ModelState
	theme      *themes.Theme
	keyMap     config.EditKeyMap

	// Recipe
	recipeID *uint
	isNew    bool
	width    int
	height   int

	// Form fields
	name             string
	description      string
	author           string
	prepTime         string
	cookTime         string
	servings         string
	url              string
	categoriesText   string
	ingredientsText  string
	instructionsText string

	// Forms
	mainForm *huh.Form
}

func NewEditModel(cookbook *db.CookBook, theme *themes.Theme, recipeID uint) (*EditModel, error) {
	cfg := config.GetGlobalConfig()
	if cfg == nil {
		return nil, fmt.Errorf("global config not set")
	}

	keymaps := cfg.Keymap.ToKeyMap().GetEditKeyMap()
	model := &EditModel{
		cookbook:   cookbook,
		keyMap:     keymaps,
		modelState: common.ModelStateLoaded,
		theme:      theme,
	}

	if recipeID != 0 {
		recipe, err := model.FetchRecipe(recipeID)
		if err != nil {
			slog.Error("Failed to fetch recipe", "error", err)
			return nil, err
		}
		model.loadRecipe(recipe)
	} else {
		model.resetRecipe()
	}

	model.setupForms()
	return model, nil
}

func (m *EditModel) Init() tea.Cmd {
	if m.mainForm == nil {
		return nil
	}
	return m.mainForm.Init()
}

func (m *EditModel) Update(msg tea.Msg) (common.TUIModel, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case messages.SaveMsg:
		cmds = append(cmds, messages.SendSessionStateMsg(common.SessionStateDetail))
		cmds = append(cmds, messages.SendRecipeSelectedMsg(msg.RecipeID))
		return m, tea.Batch(cmds...)

	case messages.EditRecipeMsg:
		if msg.Recipe == nil {
			m.resetRecipe()
		} else {
			m.loadRecipe(msg.Recipe)
		}
		m.setupForms()
		if initCmd := m.Init(); initCmd != nil {
			cmds = append(cmds, initCmd)
		}
	}

	form, huhCmd := m.mainForm.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.mainForm = f
		if huhCmd != nil {
			cmds = append(cmds, func() tea.Msg { return huhCmd() })
		}
	}

	if m.mainForm.State == huh.StateCompleted {
		save := m.mainForm.GetBool("save")
		if save {
			saveMsg, err := m.saveRecipe()
			if err != nil {
				slog.Error("Failed to save recipe", "error", err)
				return m, nil
			}
			cmds = append(cmds, messages.CmdHandler(saveMsg))
		} else {
			cmds = append(cmds, m.cancelCmds()...)
		}
	}

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
	content.WriteString(m.mainForm.View())
	return content.String()
}

func (m *EditModel) FetchRecipe(recipeID uint) (*utils.RecipeRaw, error) {
	recipe, err := m.cookbook.GetFullRecipe(recipeID)
	if err != nil {
		slog.Error("Failed to fetch recipe", "error", err)
		return nil, err
	}
	return recipe, nil
}

func (m *EditModel) resetRecipe() {
	m.recipeID = nil
	m.isNew = true
	m.name = ""
	m.description = ""
	m.author = ""
	m.prepTime = ""
	m.cookTime = ""
	m.servings = ""
	m.url = ""
	m.categoriesText = ""
	m.ingredientsText = ""
	m.instructionsText = ""
}

func (m *EditModel) loadRecipe(recipe *utils.RecipeRaw) {
	m.recipeID = &recipe.RecipeID
	m.isNew = false
	m.name = recipe.RecipeName
	m.description = recipe.RecipeDescription
	m.author = recipe.Metadata.Author
	m.prepTime = formatDurationInput(recipe.Metadata.PrepTime)
	m.cookTime = formatDurationInput(recipe.Metadata.CookTime)
	m.servings = recipe.Metadata.Quantity
	m.url = recipe.Metadata.URL
	m.categoriesText = categoriesToText(recipe.Metadata.Categories)
	m.ingredientsText = ingredientsToText(recipe.Metadata.Ingredients)
	m.instructionsText = instructionsToText(recipe.Metadata.Instructions)
}

func (m *EditModel) extractFormRecipe() (*utils.RecipeRaw, error) {
	prepTime, err := parseDurationInput(m.prepTime)
	if err != nil {
		return nil, err
	}
	cookTime, err := parseDurationInput(m.cookTime)
	if err != nil {
		return nil, err
	}
	ingredients, err := parseIngredientsText(m.ingredientsText)
	if err != nil {
		return nil, err
	}
	instructions, err := parseInstructionsText(m.instructionsText)
	if err != nil {
		return nil, err
	}

	recipe := &utils.RecipeRaw{
		RecipeName:        strings.TrimSpace(m.name),
		RecipeDescription: strings.TrimSpace(m.description),
		Metadata: utils.RecipeMetadata{
			Author:       strings.TrimSpace(m.author),
			PrepTime:     prepTime,
			CookTime:     cookTime,
			TotalTime:    prepTime + cookTime,
			Quantity:     strings.TrimSpace(m.servings),
			URL:          strings.TrimSpace(m.url),
			Categories:   parseCategories(m.categoriesText),
			Ingredients:  ingredients,
			Instructions: instructions,
		},
	}

	if m.recipeID != nil {
		recipe.RecipeID = *m.recipeID
	}

	return recipe, nil
}

func (m *EditModel) saveRecipe() (tea.Msg, error) {
	recipe, err := m.extractFormRecipe()
	if err != nil {
		return nil, err
	}

	var recipeID uint
	if m.isNew {
		recipeID, err = m.cookbook.SaveScrapedRecipe(recipe)
		if err != nil {
			return nil, err
		}
		m.recipeID = &recipeID
		m.isNew = false
	} else {
		if err := m.cookbook.UpdateRecipe(recipe); err != nil {
			return nil, err
		}
		recipeID = recipe.RecipeID
	}

	return messages.SaveMsg{RecipeID: recipeID}, nil
}

func (m *EditModel) cancelCmds() []tea.Cmd {
	if m.recipeID != nil {
		return []tea.Cmd{
			messages.SendSessionStateMsg(common.SessionStateDetail),
			messages.SendRecipeSelectedMsg(*m.recipeID),
		}
	}

	return []tea.Cmd{messages.SendSessionStateMsg(common.SessionStateList)}
}

func (m *EditModel) setupForms() {
	allAuthors, err := m.cookbook.GetAllAuthors()
	if err != nil {
		slog.Error("Failed to get all authors", "error", err)
	}

	categoryHint := "Comma-separated categories, for example: dinner, pasta, weeknight"
	if allCategories, err := m.cookbook.GetAllCategories(); err == nil && len(allCategories) > 0 {
		categoryHint = fmt.Sprintf("Comma-separated categories. Existing: %s", strings.Join(allCategories, ", "))
	} else if err != nil {
		slog.Error("Failed to get all categories", "error", err)
	}

	width := 80
	if m.width > 4 {
		width = m.width - 4
	}

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
				Description("Add a short description or notes about the recipe").
				Value(&m.description).
				Lines(4).
				Placeholder("A cozy pasta you can get on the table in 30 minutes."),

			huh.NewInput().
				Key("author").
				Title("Author").
				Description("Who created this recipe?").
				Value(&m.author).
				Suggestions(allAuthors),

			huh.NewInput().
				Key("prepTime").
				Title("Prep Time").
				Description("Optional. Use formats like 15m, 1h, or 1h 30m.").
				Value(&m.prepTime).
				Validate(validateDurationField).
				Placeholder("15m"),

			huh.NewInput().
				Key("cookTime").
				Title("Cook Time").
				Description("Optional. Use formats like 15m, 1h, or 1h 30m.").
				Value(&m.cookTime).
				Validate(validateDurationField).
				Placeholder("20m"),

			huh.NewInput().
				Key("servings").
				Title("Servings").
				Description("Optional serving yield, for example 4 servings or 2 loaves.").
				Value(&m.servings).
				Placeholder("4 servings"),

			huh.NewInput().
				Key("url").
				Title("Recipe URL").
				Description("Optional source URL.").
				Value(&m.url).
				Placeholder("https://example.com/recipe").
				Validate(validateOptionalURL),

			huh.NewInput().
				Key("categories").
				Title("Categories").
				Description(categoryHint).
				Value(&m.categoriesText).
				Placeholder("dinner, pasta, vegetarian"),

			huh.NewText().
				Key("ingredients").
				Title("Ingredients").
				Description("One ingredient per line. Add a group heading with a line ending in : like 'For the sauce:'").
				Value(&m.ingredientsText).
				Lines(10).
				Placeholder("12 ounce spaghetti\n2 tbsp olive oil\n4 cloves garlic (minced)\nFor serving:\n1/2 cup parmesan (grated)").
				Validate(validateIngredientsField),

			huh.NewText().
				Key("instructions").
				Title("Instructions").
				Description("One step per line. Numbering is optional.").
				Value(&m.instructionsText).
				Lines(10).
				Placeholder("Boil the pasta in salted water.\nWarm the oil and cook the garlic for 1 minute.\nToss everything together and serve.").
				Validate(validateInstructionsField),

			huh.NewConfirm().
				Key("save").
				Title("Save").
				Description("Save this recipe now?").
				Affirmative("Yes").
				Negative("No"),
		),
	).
		WithTheme(huh.ThemeFunc(huh.ThemeCharm)).
		WithWidth(width)
}

func (m *EditModel) GetModelState() common.ModelState {
	return m.modelState
}

func (m *EditModel) GetSessionState() common.SessionState {
	return common.SessionStateEdit
}

func (m *EditModel) GetSize() (width int, height int) {
	return m.width, m.height
}

func (m *EditModel) SetSize(width, height int) {
	m.width = width
	m.height = height

	if m.mainForm != nil {
		m.mainForm = m.mainForm.WithWidth(max(20, width-4))
	}
}

func (m *EditModel) GetCurrentTheme() *themes.Theme {
	return m.theme
}

func (m *EditModel) SetTheme(theme *themes.Theme) {
	m.theme = theme
	if m.mainForm != nil {
		m.setupForms()
	}
}
