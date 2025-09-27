package list

import (
	"fmt"

	"github.com/GarroshIcecream/yummy/yummy/config"
	db "github.com/GarroshIcecream/yummy/yummy/db"
	"github.com/GarroshIcecream/yummy/yummy/recipe"
	styles "github.com/GarroshIcecream/yummy/yummy/tui/styles"
	"github.com/GarroshIcecream/yummy/yummy/ui"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type ListModel struct {
	cookbook        *db.CookBook
	err             error
	RecipeList      list.Model
	modelState      ui.ModelState
	filterFavourite bool
	width           int
	height          int
	keyMap          config.KeyMap
}

func New(cookbook *db.CookBook, keymaps config.KeyMap, filterFavourite bool) *ListModel {
	recipes, err := cookbook.AllRecipes(filterFavourite)

	var items []list.Item
	for _, recipe := range recipes {
		items = append(items, recipe)
	}

	d := list.NewDefaultDelegate()
	d.Styles = styles.GetDelegateStyles()

	l := list.New(items, d, 80, 40)
	l.Styles = styles.GetListStyles()
	l.Title = ui.ListTitle
	l.KeyMap = keymaps.ListKeyMap()
	l.SetStatusBarItemName(ui.ListItemNameSingular, ui.ListItemNamePlural)
	l.StatusMessageLifetime = ui.ListViewStatusMessageTTL
	l.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{keymaps.Add, keymaps.Delete}
	}

	l.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{keymaps.Add, keymaps.Delete, keymaps.SetFavourite}
	}

	return &ListModel{
		cookbook:        cookbook,
		keyMap:          keymaps,
		filterFavourite: filterFavourite,
		modelState:      ui.ModelStateLoaded,
		err:             err,
		RecipeList:      l,
	}
}

func (m *ListModel) Init() tea.Cmd {
	return nil
}

func (m *ListModel) SelectedItemToRecipeWithDescription() (recipe.RecipeWithDescription, bool) {
	if len(m.RecipeList.Items()) == 0 {
		return recipe.RecipeWithDescription{}, false
	}

	if i, ok := m.RecipeList.SelectedItem().(recipe.RecipeWithDescription); ok {
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

	case ui.SetFavouriteMsg:
		if err := m.cookbook.SetFavourite(msg.RecipeID); err != nil {
			m.err = err
			return m, nil
		}
		cmd = m.RefreshRecipeList()
		cmds = append(cmds, cmd)
		isFavourite := m.RecipeList.SelectedItem().(recipe.RecipeWithDescription).IsFavourite
		if isFavourite {
			cmds = append(cmds, m.RecipeList.NewStatusMessage(ui.ListViewStatusMessageFavouriteSet))
		} else {
			cmds = append(cmds, m.RecipeList.NewStatusMessage(ui.ListViewStatusMessageFavouriteRemoved))
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
					cmds = append(cmds, m.RecipeList.NewStatusMessage(ui.ListViewStatusMessageRecipeDeleted))
				}
			}
		case key.Matches(msg, m.keyMap.Enter):
			if m.RecipeList.FilterState() != list.Filtering {
				if i, ok := m.SelectedItemToRecipeWithDescription(); ok {
					cmds = append(cmds, ui.SendSessionStateMsg(ui.SessionStateDetail))
					cmds = append(cmds, ui.SendRecipeSelectedMsg(i.RecipeID))
				}
			}

		case key.Matches(msg, m.keyMap.SetFavourite):
			if m.RecipeList.FilterState() != list.Filtering {
				if i, ok := m.SelectedItemToRecipeWithDescription(); ok {
					cmds = append(cmds, ui.SendSetFavouriteMsg(i.RecipeID))
				}
			}
		}
	case tea.WindowSizeMsg:
		h, v := styles.DocStyle.GetFrameSize()
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

	return styles.DocStyle.Render(m.RecipeList.View())
}

func (m *ListModel) RefreshRecipeList() tea.Cmd {
	recipes, err := m.cookbook.AllRecipes(m.filterFavourite)
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
		h, v := styles.DocStyle.GetFrameSize()
		m.RecipeList.SetSize(width-h, height-v)
	}
}

// GetSize returns the current width and height of the model
func (m *ListModel) GetSize() (width, height int) {
	return m.width, m.height
}

func (m *ListModel) GetModelState() ui.ModelState {
	return m.modelState
}
