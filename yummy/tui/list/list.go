package list

import (
	"fmt"
	"time"

	"github.com/GarroshIcecream/yummy/yummy/config"
	consts "github.com/GarroshIcecream/yummy/yummy/consts"
	db "github.com/GarroshIcecream/yummy/yummy/db"
	messages "github.com/GarroshIcecream/yummy/yummy/models/msg"
	"github.com/GarroshIcecream/yummy/yummy/recipe"
	themes "github.com/GarroshIcecream/yummy/yummy/themes"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type ListModel struct {
	cookbook   *db.CookBook
	err        error
	RecipeList list.Model
	modelState consts.ModelState
	config     *config.ListConfig
	width      int
	height     int
	keyMap     config.KeyMap
	theme      *themes.Theme
}

func New(cookbook *db.CookBook, keymaps config.KeyMap, theme *themes.Theme, config *config.ListConfig) *ListModel {
	recipes, err := cookbook.AllRecipes()

	var items []list.Item
	for _, recipe := range recipes {
		items = append(items, recipe)
	}

	d := list.NewDefaultDelegate()
	d.Styles = theme.DelegateStyles

	l := list.New(items, d, 80, 40)
	l.Styles = theme.ListStyles
	l.Title = config.Title
	l.KeyMap = keymaps.ListKeyMap()
	l.SetStatusBarItemName(config.ItemNameSingular, config.ItemNamePlural)
	l.StatusMessageLifetime = time.Duration(config.ViewStatusMessageTTL) * time.Millisecond
	l.Filter = CustomFilter
	l.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{keymaps.Add, keymaps.Delete}
	}

	l.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{keymaps.Add, keymaps.Delete, keymaps.SetFavourite}
	}

	return &ListModel{
		cookbook:   cookbook,
		keyMap:     keymaps,
		config:     config,
		modelState: consts.ModelStateLoaded,
		err:        err,
		RecipeList: l,
		theme:      theme,
	}
}

func (m *ListModel) Init() tea.Cmd {
	return nil
}

func (m *ListModel) SelectedItemToRecipeWithDescription() (recipe.RecipeWithDescription, bool) {
	if len(m.RecipeList.Items()) == 0 {
		return recipe.RecipeWithDescription{}, false
	}

	selectedItem := m.RecipeList.SelectedItem()
	if selectedItem == nil {
		return recipe.RecipeWithDescription{}, false
	}

	if i, ok := selectedItem.(recipe.RecipeWithDescription); ok {
		return i, true
	}

	return recipe.RecipeWithDescription{}, false
}

func (m *ListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {

	case messages.SetFavouriteMsg:
		newFavourite, err := m.cookbook.SetFavourite(msg.RecipeID)
		if err != nil {
			m.err = err
			return m, nil
		}
		cmd = m.RefreshRecipeList()
		cmds = append(cmds, cmd)
		if newFavourite {
			cmds = append(cmds, m.RecipeList.NewStatusMessage(m.config.ViewStatusMessageFavouriteSet))
		} else {
			cmds = append(cmds, m.RecipeList.NewStatusMessage(m.config.ViewStatusMessageFavouriteRemoved))
		}

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keyMap.Delete):
			if m.RecipeList.FilterState() != list.Filtering {
				if i, ok := m.SelectedItemToRecipeWithDescription(); ok {
					if err := m.cookbook.DeleteRecipe(i.RecipeID); err != nil {
						m.err = err
						return m, nil
					}

					cmd = m.RefreshRecipeList()
					cmds = append(cmds, cmd)
					cmds = append(cmds, m.RecipeList.NewStatusMessage(m.config.ViewStatusMessageRecipeDeleted))
				}
			}
		case key.Matches(msg, m.keyMap.Enter):
			if m.RecipeList.FilterState() != list.Filtering {
				if i, ok := m.SelectedItemToRecipeWithDescription(); ok {
					cmds = append(cmds, messages.SendSessionStateMsg(consts.SessionStateDetail))
					cmds = append(cmds, messages.SendRecipeSelectedMsg(i.RecipeID))
				}
			}

		case key.Matches(msg, m.keyMap.SetFavourite):
			if m.RecipeList.FilterState() != list.Filtering {
				if i, ok := m.SelectedItemToRecipeWithDescription(); ok {
					cmds = append(cmds, messages.SendSetFavouriteMsg(i.RecipeID))
				}
			}
		}
	case tea.WindowSizeMsg:
		h, v := m.theme.Doc.GetFrameSize()
		m.RecipeList.SetSize(msg.Width-h, msg.Height-v)
	}

	m.RecipeList, cmd = m.RecipeList.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Sequence(cmds...)
}

func (m *ListModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v", m.err)
	}

	return m.theme.Doc.Render(m.RecipeList.View())
}

func (m *ListModel) RefreshRecipeList() tea.Cmd {
	recipes, err := m.cookbook.AllRecipes()
	if err != nil {
		m.err = err
		return nil
	}

	var items []list.Item
	for _, recipe := range recipes {
		items = append(items, recipe)
	}

	cmd := m.RecipeList.SetItems(items)
	m.RecipeList.ResetSelected()
	return cmd
}

// SetSize sets the width and height of the model
func (m *ListModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	if m.RecipeList.Width() != 0 || m.RecipeList.Height() != 0 {
		h, v := m.theme.Doc.GetFrameSize()
		m.RecipeList.SetSize(width-h, height-v)
	}
}

// GetSize returns the current width and height of the model
func (m *ListModel) GetSize() (width, height int) {
	return m.width, m.height
}

func (m *ListModel) GetModelState() consts.ModelState {
	return m.modelState
}

func (m *ListModel) GetSessionState() consts.SessionState {
	return consts.SessionStateList
}
