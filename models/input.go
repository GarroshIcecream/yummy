package models

import (
	"fmt"

	db "github.com/GarroshIcecream/yummy/db"
	keys "github.com/GarroshIcecream/yummy/keymaps"
	recipes "github.com/GarroshIcecream/yummy/recipe"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type RecipeLoadedMsg struct {
	recipe_id uint
	err       error
}

type InputModel struct {
	cookbook  *db.CookBook
	err       error
	inputURL  textinput.Model
	spinner   spinner.Model
	isLoading bool
}

func NewInputModel(cookbook db.CookBook, url *string) *InputModel {
	inputURL := textinput.New()
	inputURL.Placeholder = "Enter Recipe URL"
	inputURL.Focus()

	if url != nil {
		inputURL.SetValue(*url)
	}

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B"))

	return &InputModel{
		cookbook:  &cookbook,
		err:       nil,
		inputURL:  inputURL,
		spinner:   s,
		isLoading: false,
	}
}

func (m *InputModel) Init() tea.Cmd {
	return nil
}

func (m *InputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, keys.Keys.Enter):
			if !m.isLoading {
				m.isLoading = true
				url := m.inputURL.Value()
				m.inputURL.Reset()
				cmds = append(cmds, m.spinner.Tick)
				cmds = append(cmds, func() tea.Msg {
					return m.handleURLInput(url)
				})
			}
		case key.Matches(msg, keys.Keys.Back):
			return NewListModel(*m.cookbook, nil), nil
		}

	case spinner.TickMsg:
		if m.isLoading {
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}

	case RecipeLoadedMsg:
		m.isLoading = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		return NewDetailModel(*m.cookbook, msg.recipe_id), nil
	}

	m.inputURL, cmd = m.inputURL.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *InputModel) View() string {
	if m.err != nil {
		return fmt.Sprintf(
			"Error: %v\n\nPress esc to go back",
			m.err,
		)
	}

	if m.isLoading {
		return fmt.Sprintf(
			"\n  %s Scraping recipe...",
			m.spinner.View(),
		)
	}

	return fmt.Sprintf(
		"Enter Recipe URL:\n\n%s\n\n%s",
		m.inputURL.View(),
		"(esc to quit)",
	) + "\n"
}

func (m *InputModel) handleURLInput(url string) tea.Msg {
	recipeRaw, err := recipes.GetRecipeFromURL(url)
	if err != nil {
		return RecipeLoadedMsg{err: err}
	}

	id, err := m.cookbook.SaveScrapedRecipe(recipeRaw)
	if err != nil {
		return RecipeLoadedMsg{err: err}
	}

	return RecipeLoadedMsg{recipe_id: id, err: nil}
}
