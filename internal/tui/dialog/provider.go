package dialog

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/GarroshIcecream/yummy/internal/config"
	"github.com/GarroshIcecream/yummy/internal/llms"
	common "github.com/GarroshIcecream/yummy/internal/models/common"
	messages "github.com/GarroshIcecream/yummy/internal/models/msg"
	themes "github.com/GarroshIcecream/yummy/internal/themes"
	"github.com/GarroshIcecream/yummy/internal/utils"
)

// ProviderAppliedMsg asks the app to install a prepared provider (the TUI
// manager applies it to the LLM service, rebuilds agents and saves config).
type ProviderAppliedMsg struct {
	Name  string
	State llms.ProviderState
}

const (
	loginTimeout  = 5 * time.Minute
	switchTimeout = 2 * time.Minute
)

// SwitchProviderCmd prepares the named provider off the UI goroutine and
// reports it as a ProviderAppliedMsg (with State.Err set if it isn't ready).
func SwitchProviderCmd(name string) tea.Cmd {
	return func() tea.Msg {
		return ProviderAppliedMsg{Name: name, State: prepareProvider(name)}
	}
}

func prepareProvider(name string) llms.ProviderState {
	cfg := config.GetChatConfig()
	cfg.Provider = name
	p, err := llms.NewProvider(cfg, config.DataDir())
	if err != nil {
		return llms.ProviderState{Err: err}
	}
	ctx, cancel := context.WithTimeout(context.Background(), switchTimeout)
	defer cancel()
	return llms.PrepareProvider(ctx, p)
}

type providerPhase int

const (
	providerPhaseChoose providerPhase = iota
	providerPhaseSigningIn
	providerPhaseSwitching
	providerPhaseDone
	providerPhaseFailed
)

type providerItem struct {
	name  string
	label string
}

var providerItems = []providerItem{
	{name: "ollama", label: "Ollama"},
	{name: "chatgpt", label: "ChatGPT"},
}

func providerLabel(name string) string {
	for _, it := range providerItems {
		if it.name == name {
			return it.label
		}
	}
	return name
}

// providerDeps are the side effects of the dialog, swapped out in tests.
type providerDeps struct {
	signedIn func() bool
	login    func(ctx context.Context, openBrowser func(url string) error) (account string, err error)
	browser  func(url string) error
	logout   func() error
	prepare  func(name string) llms.ProviderState
}

func defaultProviderDeps() providerDeps {
	return providerDeps{
		signedIn: func() bool { return llms.ChatGPTSignedIn(config.DataDir()) },
		login: func(ctx context.Context, openBrowser func(string) error) (string, error) {
			t, err := llms.ChatGPTLogin(ctx, config.DataDir(), openBrowser)
			if err != nil {
				return "", err
			}
			return t.AccountID, nil
		},
		browser: utils.OpenBrowser,
		logout:  func() error { return llms.ChatGPTLogout(config.DataDir()) },
		prepare: prepareProvider,
	}
}

type providerLoginURLMsg struct{ url string }

type providerLoginResultMsg struct {
	account string
	err     error
}

type providerPreparedMsg struct {
	name  string
	state llms.ProviderState
}

// ProviderDialogCmp lets the user pick the LLM provider, signs in to
// ChatGPT when needed, and switches the running app over.
type ProviderDialogCmp struct {
	theme    *themes.Theme
	spinner  spinner.Model
	width    int
	height   int
	deps     providerDeps
	phase    providerPhase
	current  string
	selected int
	signedIn bool

	target   string // provider being switched to
	authURL  string
	account  string
	errMsg   string
	doneText string
	notice   string // one-line status under the chooser
	cancel   context.CancelFunc
}

// NewProviderDialog opens the provider chooser; current is the active
// provider's name.
func NewProviderDialog(current string, theme *themes.Theme) *ProviderDialogCmp {
	return newProviderDialog(current, theme, defaultProviderDeps())
}

func newProviderDialog(current string, theme *themes.Theme, deps providerDeps) *ProviderDialogCmp {
	s := spinner.New()
	s.Spinner = spinner.Dot
	d := &ProviderDialogCmp{
		theme:    theme,
		spinner:  s,
		width:    64,
		deps:     deps,
		current:  current,
		signedIn: deps.signedIn(),
	}
	for i, it := range providerItems {
		if it.name == current {
			d.selected = i
		}
	}
	return d
}

func (d *ProviderDialogCmp) Init() tea.Cmd {
	if d.phase == providerPhaseSigningIn {
		return d.startLogin()
	}
	return nil
}

func (d *ProviderDialogCmp) startLogin() tea.Cmd {
	d.phase = providerPhaseSigningIn
	d.target = "chatgpt"
	d.authURL, d.errMsg = "", ""
	ctx, cancel := context.WithTimeout(context.Background(), loginTimeout)
	d.cancel = cancel
	urls := make(chan string, 1)
	deps := d.deps
	return tea.Batch(
		d.spinner.Tick,
		func() tea.Msg {
			account, err := deps.login(ctx, func(u string) error {
				select {
				case urls <- u:
				default:
				}
				return deps.browser(u)
			})
			return providerLoginResultMsg{account: account, err: err}
		},
		func() tea.Msg {
			select {
			case u := <-urls:
				return providerLoginURLMsg{url: u}
			case <-ctx.Done():
				return nil
			}
		},
	)
}

func (d *ProviderDialogCmp) startSwitch(name string) tea.Cmd {
	d.phase = providerPhaseSwitching
	d.target = name
	d.errMsg = ""
	prepare := d.deps.prepare
	return tea.Batch(d.spinner.Tick, func() tea.Msg {
		return providerPreparedMsg{name: name, state: prepare(name)}
	})
}

// signOut removes the ChatGPT session. When ChatGPT is the active provider
// it is re-checked so the app reports that it is no longer ready.
func (d *ProviderDialogCmp) signOut() tea.Cmd {
	if err := d.deps.logout(); err != nil {
		d.notice = "Couldn't sign out: " + err.Error()
		return nil
	}
	d.signedIn = false
	d.notice = "Signed out of ChatGPT."
	if d.current != "chatgpt" {
		return nil
	}
	prepare := d.deps.prepare
	return func() tea.Msg {
		return ProviderAppliedMsg{Name: "chatgpt", State: prepare("chatgpt")}
	}
}

func (d *ProviderDialogCmp) close() tea.Cmd {
	if d.cancel != nil {
		d.cancel()
	}
	return messages.SendCloseModalViewMsg()
}

func (d *ProviderDialogCmp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case spinner.TickMsg:
		if d.phase != providerPhaseSigningIn && d.phase != providerPhaseSwitching {
			return d, nil
		}
		var cmd tea.Cmd
		d.spinner, cmd = d.spinner.Update(msg)
		return d, cmd

	case providerLoginURLMsg:
		d.authURL = msg.url
		return d, nil

	case providerLoginResultMsg:
		if d.phase != providerPhaseSigningIn {
			return d, nil
		}
		if msg.err != nil {
			d.phase = providerPhaseFailed
			d.errMsg = "Sign-in failed: " + msg.err.Error()
			if errors.Is(msg.err, context.DeadlineExceeded) {
				d.errMsg = "Sign-in timed out. Try again."
			}
			return d, nil
		}
		d.account = msg.account
		d.signedIn = true
		return d, d.startSwitch("chatgpt")

	case providerPreparedMsg:
		if d.phase != providerPhaseSwitching || msg.name != d.target {
			return d, nil
		}
		if msg.state.Err != nil {
			d.phase = providerPhaseFailed
			d.errMsg = fmt.Sprintf("Couldn't switch to %s: %v", providerLabel(msg.name), msg.state.Err)
			return d, nil
		}
		d.phase = providerPhaseDone
		d.doneText = fmt.Sprintf("Now using %s · %s", providerLabel(msg.name), msg.state.ModelName)
		d.current = msg.name
		applied := ProviderAppliedMsg{Name: msg.name, State: msg.state}
		return d, func() tea.Msg { return applied }

	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc", "q":
			return d, d.close()
		}
		switch d.phase {
		case providerPhaseChoose:
			switch msg.String() {
			case "up", "ctrl+k", "k":
				if d.selected > 0 {
					d.selected--
				}
			case "down", "ctrl+j", "j":
				if d.selected < len(providerItems)-1 {
					d.selected++
				}
			case "x":
				if providerItems[d.selected].name == "chatgpt" && d.signedIn {
					return d, d.signOut()
				}
			case "enter":
				d.notice = ""
				name := providerItems[d.selected].name
				if name == "chatgpt" && !d.signedIn {
					return d, d.startLogin()
				}
				return d, d.startSwitch(name)
			}
		case providerPhaseDone:
			if msg.String() == "enter" {
				return d, d.close()
			}
		case providerPhaseFailed:
			if msg.String() == "enter" {
				d.phase = providerPhaseChoose
				d.errMsg = ""
			}
		}
	}
	return d, nil
}

func (d *ProviderDialogCmp) View() tea.View {
	innerWidth := max(d.width-6, 40)
	th := d.theme
	row := func(s lipgloss.Style, text string) string {
		return s.Width(innerWidth).Align(lipgloss.Center).Render(text)
	}
	keyHelp := func(pairs ...string) string {
		var parts []string
		for i := 0; i+1 < len(pairs); i += 2 {
			parts = append(parts, th.AddRecipeFromURLKeyHighlight.Render(pairs[i])+th.AddRecipeFromURLHelp.Render(" "+pairs[i+1]))
		}
		return lipgloss.NewStyle().Width(innerWidth).Align(lipgloss.Center).
			Render(strings.Join(parts, th.AddRecipeFromURLHelp.Render(" | ")))
	}
	spin := func(text string) string {
		return lipgloss.NewStyle().Width(innerWidth).Align(lipgloss.Center).
			Render(th.AddRecipeFromURLSpinner.Render(d.spinner.View()) + th.AddRecipeFromURLPrompt.Render(" "+text))
	}

	title := row(th.AddRecipeFromURLTitle, "LLM Provider")
	sep := th.AddRecipeFromURLSeparator.Render(strings.Repeat("─", innerWidth))

	var body []string
	var help string
	switch d.phase {
	case providerPhaseChoose:
		for i, it := range providerItems {
			marker := "  "
			if it.name == d.current {
				marker = "* "
			}
			detail := "local models"
			if it.name == "chatgpt" {
				detail = "your Plus/Pro subscription · not signed in"
				if d.signedIn {
					detail = "your Plus/Pro subscription · signed in"
				}
			}
			line := fmt.Sprintf("%s%-8s %s", marker, it.label, detail)
			style := th.DialogUnselectedRow
			if i == d.selected {
				style = th.DialogSelectedRow
			}
			body = append(body, style.Width(innerWidth).Render(line))
		}
		if d.notice != "" {
			body = append(body, "", row(th.AddRecipeFromURLHelp, d.notice))
		}
		enterHelp := "use"
		if providerItems[d.selected].name == "chatgpt" && !d.signedIn {
			enterHelp = "sign in & use"
		}
		pairs := []string{"↑/↓", "choose", "enter", enterHelp}
		if providerItems[d.selected].name == "chatgpt" && d.signedIn {
			pairs = append(pairs, "x", "sign out")
		}
		help = keyHelp(append(pairs, "esc", "cancel")...)

	case providerPhaseSigningIn:
		body = append(body, spin("Waiting for ChatGPT sign-in in your browser…"))
		if d.authURL != "" {
			body = append(body, "", row(th.AddRecipeFromURLHelp, "If the browser didn't open, visit:"),
				th.AddRecipeFromURLPrompt.Width(innerWidth).Render(utils.WrapTextToWidth(d.authURL, innerWidth)))
		}
		help = keyHelp("esc", "cancel")

	case providerPhaseSwitching:
		text := "Switching to " + providerLabel(d.target) + "…"
		if d.account != "" {
			text = "Signed in. " + text
		}
		body = append(body, spin(text))
		help = keyHelp("esc", "close")

	case providerPhaseDone:
		body = append(body, row(th.AddRecipeFromURLPrompt, d.doneText))
		help = keyHelp("enter", "close")

	case providerPhaseFailed:
		body = append(body, th.AddRecipeFromURLError.Width(innerWidth).Align(lipgloss.Center).
			Render(utils.WrapTextToWidth(d.errMsg, innerWidth)))
		help = keyHelp("enter", "back", "esc", "close")
	}

	parts := append([]string{title, sep, ""}, body...)
	parts = append(parts, "", help)
	content := utils.FillBackground(lipgloss.JoinVertical(lipgloss.Left, parts...), th.SurfaceBg, innerWidth)
	rendered := th.AddRecipeFromURLDialog.Width(d.width).Render(content)
	return tea.NewView(th.AddRecipeFromURLContainer.Render(rendered))
}

// Busy reports whether a sign-in or switch is in progress.
func (d *ProviderDialogCmp) Busy() (bool, string) {
	switch d.phase {
	case providerPhaseSigningIn:
		return true, "Signing in to ChatGPT"
	case providerPhaseSwitching:
		return true, "Switching LLM provider"
	}
	return false, ""
}

func (d *ProviderDialogCmp) SetSize(width, height int) {
	if width > 0 {
		d.width = min(clampModalWidth(width, d.width, 44), 80)
	}
	d.height = clampModalHeight(height, d.height, 10)
}

func (d *ProviderDialogCmp) GetSize() (int, int) { return d.width, d.height }

func (d *ProviderDialogCmp) GetModelState() common.ModelState { return common.ModelStateLoaded }
