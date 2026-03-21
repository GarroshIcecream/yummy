package dialog

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	common "github.com/GarroshIcecream/yummy/internal/models/common"
	messages "github.com/GarroshIcecream/yummy/internal/models/msg"
	"github.com/GarroshIcecream/yummy/internal/scrape"
	themes "github.com/GarroshIcecream/yummy/internal/themes"
)

type scraperUpdatePhase int

const (
	scraperPhaseChecking  scraperUpdatePhase = iota // fetching installed + latest
	scraperPhaseReady                               // showing comparison
	scraperPhaseUpgrading                           // pip running
	scraperPhaseDone                                // upgrade complete
)

type scraperVersionsReadyMsg struct {
	installed string
	latest    string
	err       error
}

type scraperUpgradeCompleteMsg struct {
	newVersion string
	err        error
}

type ScraperUpdateDialogCmp struct {
	theme      *themes.Theme
	pythonPath string
	spinner    spinner.Model
	width      int
	height     int
	phase      scraperUpdatePhase
	installed  string
	latest     string
	errMsg     string
}

func NewScraperUpdateDialog(pythonPath string, theme *themes.Theme) *ScraperUpdateDialogCmp {
	s := spinner.New()
	s.Spinner = spinner.Dot
	return &ScraperUpdateDialogCmp{
		theme:      theme,
		pythonPath: pythonPath,
		spinner:    s,
		width:      60,
		phase:      scraperPhaseChecking,
	}
}

func (m *ScraperUpdateDialogCmp) Init() tea.Cmd {
	pythonPath := m.pythonPath
	return tea.Batch(
		m.spinner.Tick,
		func() tea.Msg {
			installed, err := scrape.InstalledVersion(pythonPath)
			if err != nil {
				return scraperVersionsReadyMsg{err: fmt.Errorf("could not detect installed version: %w", err)}
			}
			latest, err := scrape.LatestVersion()
			if err != nil {
				return scraperVersionsReadyMsg{installed: installed, err: fmt.Errorf("could not reach PyPI: %w", err)}
			}
			return scraperVersionsReadyMsg{installed: installed, latest: latest}
		},
	)
}

func (m *ScraperUpdateDialogCmp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case scraperVersionsReadyMsg:
		if msg.err != nil {
			m.errMsg = msg.err.Error()
			if len(m.errMsg) > 70 {
				m.errMsg = m.errMsg[:67] + "..."
			}
			m.phase = scraperPhaseReady
			return m, nil
		}
		m.installed = msg.installed
		m.latest = msg.latest
		m.phase = scraperPhaseReady
		return m, nil

	case scraperUpgradeCompleteMsg:
		if msg.err != nil {
			m.errMsg = msg.err.Error()
			if len(m.errMsg) > 70 {
				m.errMsg = m.errMsg[:67] + "..."
			}
		} else {
			m.installed = msg.newVersion
		}
		m.phase = scraperPhaseDone
		return m, nil

	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc", "q":
			return m, messages.SendCloseModalViewMsg()

		case "enter", "u":
			if m.phase == scraperPhaseReady && m.errMsg == "" && m.installed != m.latest {
				m.phase = scraperPhaseUpgrading
				pythonPath := m.pythonPath
				return m, tea.Batch(
					m.spinner.Tick,
					func() tea.Msg {
						if err := scrape.UpgradeRecipeScrapers(pythonPath); err != nil {
							return scraperUpgradeCompleteMsg{err: err}
						}
						newVersion, err := scrape.InstalledVersion(pythonPath)
						if err != nil {
							return scraperUpgradeCompleteMsg{err: err}
						}
						return scraperUpgradeCompleteMsg{newVersion: newVersion}
					},
				)
			}
			if m.phase == scraperPhaseDone {
				newVersion := m.installed
				return m, tea.Batch(
					messages.SendCloseModalViewMsg(),
					func() tea.Msg {
						return messages.RecipeScrapersVersionMsg{Version: newVersion, Latest: newVersion}
					},
				)
			}
		}
	}

	return m, nil
}

func (m *ScraperUpdateDialogCmp) View() tea.View {
	innerWidth := m.width - 6
	if innerWidth < 40 {
		innerWidth = 40
	}

	dialogBorder := lipgloss.Border{
		Top:         "▀",
		Bottom:      "▄",
		Left:        "▌",
		Right:       "▐",
		TopLeft:     "▛",
		TopRight:    "▜",
		BottomLeft:  "▙",
		BottomRight: "▟",
	}

	row := func(s lipgloss.Style, align lipgloss.Position, text string) string {
		return s.Width(innerWidth).Align(align).Render(text)
	}

	title := row(m.theme.AddRecipeFromURLTitle, lipgloss.Center, "recipe-scrapers Update")
	sep := m.theme.AddRecipeFromURLSeparator.Render(strings.Repeat("─", innerWidth))

	var body, help string

	switch m.phase {
	case scraperPhaseChecking:
		body = lipgloss.NewStyle().Width(innerWidth).Align(lipgloss.Center).Render(
			m.theme.AddRecipeFromURLSpinner.Render(m.spinner.View()) +
				m.theme.AddRecipeFromURLPrompt.Render(" Checking versions…"),
		)
		help = row(m.theme.AddRecipeFromURLHelp, lipgloss.Center, "esc  cancel")

	case scraperPhaseReady:
		if m.errMsg != "" {
			body = row(m.theme.AddRecipeFromURLError, lipgloss.Center, m.errMsg)
			help = row(m.theme.AddRecipeFromURLHelp, lipgloss.Center, "esc  close")
		} else if m.installed == m.latest {
			body = row(m.theme.AddRecipeFromURLPrompt, lipgloss.Center,
				fmt.Sprintf("Already on the latest version (v%s)", m.installed))
			help = row(m.theme.AddRecipeFromURLHelp, lipgloss.Center, "esc  close")
		} else {
			body = lipgloss.JoinVertical(lipgloss.Center,
				row(m.theme.AddRecipeFromURLPrompt, lipgloss.Center,
					fmt.Sprintf("Installed:  v%s", m.installed)),
				row(m.theme.AddRecipeFromURLPrompt, lipgloss.Center,
					fmt.Sprintf("Available:  v%s", m.latest)),
			)
			enterKey := m.theme.AddRecipeFromURLKeyHighlight.Render("enter")
			escKey := m.theme.AddRecipeFromURLKeyHighlight.Render("esc")
			dot := m.theme.AddRecipeFromURLHelp.Render(" · ")
			help = lipgloss.NewStyle().Width(innerWidth).Align(lipgloss.Center).Render(
				enterKey + m.theme.AddRecipeFromURLHelp.Render(" update") +
					dot + escKey + m.theme.AddRecipeFromURLHelp.Render(" cancel"),
			)
		}

	case scraperPhaseUpgrading:
		body = lipgloss.NewStyle().Width(innerWidth).Align(lipgloss.Center).Render(
			m.theme.AddRecipeFromURLSpinner.Render(m.spinner.View()) +
				m.theme.AddRecipeFromURLPrompt.Render(" Updating…"),
		)
		help = row(m.theme.AddRecipeFromURLHelp, lipgloss.Center, "please wait")

	case scraperPhaseDone:
		if m.errMsg != "" {
			body = row(m.theme.AddRecipeFromURLError, lipgloss.Center, m.errMsg)
		} else {
			body = row(m.theme.AddRecipeFromURLPrompt, lipgloss.Center,
				fmt.Sprintf("Updated to v%s", m.installed))
		}
		enterKey := m.theme.AddRecipeFromURLKeyHighlight.Render("enter")
		help = lipgloss.NewStyle().Width(innerWidth).Align(lipgloss.Center).Render(
			enterKey + m.theme.AddRecipeFromURLHelp.Render(" close"),
		)
	}

	parts := []string{title, sep, "", body, "", help}
	content := lipgloss.JoinVertical(lipgloss.Left, parts...)

	rendered := lipgloss.NewStyle().
		Border(dialogBorder).
		BorderForeground(m.theme.AddRecipeFromURLAccent.GetForeground()).
		Padding(1, 2).
		Width(m.width).
		Render(content)

	return tea.NewView(m.theme.AddRecipeFromURLContainer.Render(rendered))
}

func (m *ScraperUpdateDialogCmp) SetSize(width, height int) {
	if width > 0 {
		m.width = clampModalWidth(width, m.width, 40)
		if m.width > 70 {
			m.width = 70
		}
	}
	m.height = clampModalHeight(height, m.height, 10)
}

func (m *ScraperUpdateDialogCmp) GetSize() (int, int) {
	return m.width, m.height
}

func (m *ScraperUpdateDialogCmp) GetModelState() common.ModelState {
	return common.ModelStateLoaded
}
