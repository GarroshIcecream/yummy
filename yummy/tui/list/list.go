package list

import (
	"log/slog"
	"os"
	"time"

	"github.com/GarroshIcecream/yummy/yummy/config"
	db "github.com/GarroshIcecream/yummy/yummy/db"
	common "github.com/GarroshIcecream/yummy/yummy/models/common"
	messages "github.com/GarroshIcecream/yummy/yummy/models/msg"
	themes "github.com/GarroshIcecream/yummy/yummy/themes"
	"github.com/GarroshIcecream/yummy/yummy/utils"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/term"
)

type ListModel struct {
	cookbook   *db.CookBook
	RecipeList list.Model
	modelState common.ModelState
	config     *config.ListConfig
	width      int
	height     int
	keyMap     config.KeyMap
	theme      *themes.Theme
}

func New(cookbook *db.CookBook, keymaps config.KeyMap, theme *themes.Theme) (*ListModel, error) {
	windowWidth, windowHeight, err := term.GetSize(os.Stdout.Fd())
	if err != nil {
		slog.Error("Failed to get terminal size", "error", err)
		return nil, err
	}

	cfg := config.GetListConfig()
	recipes, err := cookbook.AllRecipes()
	if err != nil {
		slog.Error("Failed to get recipes", "error", err)
		return nil, err
	}

	var items []list.Item
	for _, recipe := range recipes {
		items = append(items, recipe)
	}

	d := list.NewDefaultDelegate()
	d.Styles = theme.DelegateStyles

	l := list.New(items, d, windowWidth, windowHeight)
	l.Styles = theme.ListStyles
	l.Title = cfg.Title
	l.KeyMap = keymaps.ListKeyMap()
	l.SetStatusBarItemName(cfg.ItemNameSingular, cfg.ItemNamePlural)
	l.StatusMessageLifetime = time.Duration(cfg.ViewStatusMessageTTL) * time.Millisecond
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
		config:     cfg,
		modelState: common.ModelStateLoaded,
		RecipeList: l,
		theme:      theme,
	}, nil
}

func (m *ListModel) Init() tea.Cmd {
	return nil
}

func (m *ListModel) SelectedItemToRecipeWithDescription() (utils.RecipeRaw, bool) {
	if len(m.RecipeList.Items()) == 0 {
		return utils.RecipeRaw{}, false
	}

	selectedItem := m.RecipeList.SelectedItem()
	if selectedItem == nil {
		return utils.RecipeRaw{}, false
	}

	if i, ok := selectedItem.(utils.RecipeRaw); ok {
		return i, true
	}

	return utils.RecipeRaw{}, false
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
			slog.Error("Failed to set favourite", "error", err)
			return m, nil
		}
		cmds = append(cmds, m.RefreshRecipeList())
		cmds = append(cmds, messages.SendFavouriteSetMsg(newFavourite))

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keyMap.Delete):
			if m.RecipeList.FilterState() != list.Filtering {
				if i, ok := m.SelectedItemToRecipeWithDescription(); ok {
					if err := m.cookbook.DeleteRecipe(i.RecipeID); err != nil {
						slog.Error("Failed to delete recipe", "error", err)
						return m, nil
					}

					cmds = append(cmds, m.RefreshRecipeList())
					cmds = append(cmds, m.RecipeList.NewStatusMessage(m.config.ViewStatusMessageRecipeDeleted))
				}
			}
		case key.Matches(msg, m.keyMap.Enter):
			if m.RecipeList.FilterState() != list.Filtering {
				if i, ok := m.SelectedItemToRecipeWithDescription(); ok {
					cmds = append(cmds, messages.SendSessionStateMsg(common.SessionStateDetail))
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
	}

	m.RecipeList, cmd = m.RecipeList.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Sequence(cmds...)
}

func (m *ListModel) View() string {
	return m.theme.Doc.Render(m.RecipeList.View())
}

func (m *ListModel) RefreshRecipeList() tea.Cmd {
	recipes, err := m.cookbook.AllRecipes()
	if err != nil {
		slog.Error("Failed to refresh recipe list", "error", err)
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
		h, v := m.theme.Doc.GetFrameSize()
		m.RecipeList.SetSize(width-h, height-v)
	}
}

// GetSize returns the current width and height of the model
func (m *ListModel) GetSize() (width, height int) {
	return m.width, m.height
}

func (m *ListModel) GetModelState() common.ModelState {
	return m.modelState
}

func (m *ListModel) GetSessionState() common.SessionState {
	return common.SessionStateList
}
