package models

import (
	"fmt"
	"log"

	db "github.com/GarroshIcecream/recipe_me/db"
	keys "github.com/GarroshIcecream/recipe_me/keymaps"
	recipes "github.com/GarroshIcecream/recipe_me/recipe"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type InputModel struct {
	cookbook *db.CookBook
	err      error
	inputURL textinput.Model
}

func NewInputModel(cookbook db.CookBook, url *string) *InputModel {
	inputURL := textinput.New()
	inputURL.Placeholder = "Enter Recipe URL"
	inputURL.Focus()

	if url != nil {
		inputURL.SetValue(*url)
	}

	return &InputModel{
		cookbook: &cookbook,
		err:      nil,
		inputURL: inputURL,
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
			recipe_id, err := m.HandleURLInput(msg)
			if err != nil {
				m.err = err
				return m, nil
			}

			return NewDetailModel(*m.cookbook, recipe_id), nil
		case key.Matches(msg, keys.Keys.Back):
			return NewListModel(*m.cookbook, nil), nil
		}
	}
	m.inputURL, cmd = m.inputURL.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *InputModel) View() string {
	if m.err != nil {
		log.Fatalf("Encountered unknown error while viewing ListModel: %s", m.err)
		return fmt.Sprintf("Error: %v", m.err)
	}

	return fmt.Sprintf(
		"Enter Recipe URL:\n\n%s\n\n%s",
		m.inputURL.View(),
		"(esc to quit)",
	) + "\n"
}

func (m *InputModel) HandleURLInput(msg tea.Msg) (uint, error) {
	url := m.inputURL.Value()
	m.inputURL.Reset()

	recipeRaw, err := recipes.GetRecipeFromURL(url)
	if err != nil {
		m.err = err
		return 0, err
	}

	id, err := m.cookbook.SaveScrapedRecipe(recipeRaw)
	if err != nil {
		m.err = err
		return 0, err
	}

	return id, nil
}
