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
	cookbook   *db.CookBook
	err        error
	RecipeList list.Model
	modelState ui.ModelState
	width      int
	height     int
	keyMap     config.KeyMap
}

func New(cookbook *db.CookBook, keymaps config.KeyMap) *ListModel {
	recipes, err := cookbook.AllRecipes()

	var items []list.Item
	for _, recipe := range recipes {
		items = append(items, recipe)
	}

	d := list.NewDefaultDelegate()
	d = styles.ApplyDelegateStyles(d)

	l := list.New(items, d, 80, 40)
	l = styles.ApplyListStyles(l)
	l.Title = "📚 My Cookbook"
	l.SetStatusBarItemName("recipe", "recipes")
	l.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{keymaps.Add, keymaps.Delete}
	}

	l.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{keymaps.Add, keymaps.Delete}
	}

	return &ListModel{
		cookbook:   cookbook,
		keyMap:     keymaps,
		modelState: ui.ModelStateLoaded,
		err:        err,
		RecipeList: l,
	}
}

func (m *ListModel) Init() tea.Cmd {
	return nil
}

func (m *ListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keyMap.Delete):
			if m.RecipeList.FilterState() != list.Filtering {
				if i, ok := m.RecipeList.SelectedItem().(recipe.RecipeWithDescription); ok {
					if err := m.cookbook.DeleteRecipe(i.RecipeID); err != nil {
						m.err = err
						return m, nil
					}

					cmd = m.RefreshRecipeList()
					cmds = append(cmds, cmd)
				}
			}
		case key.Matches(msg, m.keyMap.Enter):
			if m.RecipeList.FilterState() != list.Filtering {
				if i, ok := m.RecipeList.SelectedItem().(recipe.RecipeWithDescription); ok {
					cmds = append(cmds, ui.SendSessionStateMsg(ui.SessionStateDetail))
					cmds = append(cmds, ui.SendRecipeSelectedMsg(i.RecipeID))
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
