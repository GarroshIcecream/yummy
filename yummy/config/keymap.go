package config

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
)

type KeyMap struct {
	CursorUp             key.Binding
	CursorDown           key.Binding
	Yes                  key.Binding
	No                   key.Binding
	Add                  key.Binding
	NewSession           key.Binding
	Back                 key.Binding
	Delete               key.Binding
	Quit                 key.Binding
	Enter                key.Binding
	Help                 key.Binding
	Edit                 key.Binding
	StateSelector        key.Binding
	SessionSelector      key.Binding
	SetFavourite         key.Binding
	PrevPage             key.Binding
	NextPage             key.Binding
	ForceQuit            key.Binding
	ShowFullHelp         key.Binding
	CloseFullHelp        key.Binding
	CancelWhileFiltering key.Binding
	AcceptWhileFiltering key.Binding
	GoToStart            key.Binding
	GoToEnd              key.Binding
	Filter               key.Binding
	ClearFilter          key.Binding
	EditIngredients      key.Binding
	EditInstructions     key.Binding
	EditAdd              key.Binding
	EditEdit             key.Binding
	EditDelete           key.Binding
}

func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.StateSelector, k.Quit}
}

func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.CursorUp, k.CursorDown},        // first column
		{k.Help, k.StateSelector, k.Quit}, // second column
	}
}

func (k KeyMap) ListKeyMap() list.KeyMap {
	l := list.KeyMap{
		CursorUp:             k.CursorUp,
		CursorDown:           k.CursorDown,
		NextPage:             k.NextPage,
		PrevPage:             k.PrevPage,
		GoToStart:            k.GoToStart,
		GoToEnd:              k.GoToEnd,
		Filter:               k.Filter,
		ClearFilter:          k.ClearFilter,
		CancelWhileFiltering: k.CancelWhileFiltering,
		AcceptWhileFiltering: k.AcceptWhileFiltering,
		ShowFullHelp:         k.ShowFullHelp,
		CloseFullHelp:        k.CloseFullHelp,
		Quit:                 k.Quit,
		ForceQuit:            k.ForceQuit,
	}

	return l
}

// DefaultKeyMap returns the default keymap with standard bindings
func DefaultKeyMap() KeyMap {
	return createKeyMap(nil)
}

// CreateKeyMap creates a keymap with optional custom bindings
func CreateKeyMap(customBindings map[string][]string) KeyMap {
	return createKeyMap(customBindings)
}

// CreateKeyMapFromConfig creates a keymap using the keymap configuration
func CreateKeyMapFromConfig(keymapConfig KeymapConfig) KeyMap {
	return createKeyMap(keymapConfig.CustomBindings)
}

// createKeyMap creates a keymap with the given custom bindings
func createKeyMap(customBindings map[string][]string) KeyMap {
	// Helper function to get keys for a binding, with fallback to defaults
	getKeys := func(bindingName string, defaultKeys []string) []string {
		if customBindings != nil {
			if customKeys, exists := customBindings[bindingName]; exists {
				return customKeys
			}
		}
		return defaultKeys
	}

	return KeyMap{
		CursorUp: key.NewBinding(
			key.WithKeys(getKeys("cursor_up", []string{"k", "up"})...),
			key.WithHelp("↑/k", "move up"),
		),
		CursorDown: key.NewBinding(
			key.WithKeys(getKeys("cursor_down", []string{"j", "down"})...),
			key.WithHelp("↓/j", "move down"),
		),
		Yes: key.NewBinding(
			key.WithKeys(getKeys("yes", []string{"y"})...),
			key.WithHelp("y", "yes"),
		),
		No: key.NewBinding(
			key.WithKeys(getKeys("no", []string{"n"})...),
			key.WithHelp("n", "no"),
		),
		Add: key.NewBinding(
			key.WithKeys(getKeys("add", []string{"ctrl+a"})...),
			key.WithHelp("ctrl+a", "add recipe"),
		),
		NewSession: key.NewBinding(
			key.WithKeys(getKeys("new_session", []string{"ctrl+a"})...),
			key.WithHelp("ctrl+a", "new session"),
		),
		Back: key.NewBinding(
			key.WithKeys(getKeys("back", []string{"esc", "q"})...),
			key.WithHelp("esc/q", "go back"),
		),
		Delete: key.NewBinding(
			key.WithKeys(getKeys("delete", []string{"ctrl+x"})...),
			key.WithHelp("ctrl+x", "delete recipe"),
		),
		Quit: key.NewBinding(
			key.WithKeys(getKeys("quit", []string{"q", "esc"})...),
			key.WithHelp("esc/q", "quit"),
		),
		Enter: key.NewBinding(
			key.WithKeys(getKeys("enter", []string{"enter"})...),
			key.WithHelp("enter", "select"),
		),
		Help: key.NewBinding(
			key.WithKeys(getKeys("help", []string{"h", "?"})...),
			key.WithHelp("h/?", "help"),
		),
		Edit: key.NewBinding(
			key.WithKeys(getKeys("edit", []string{"ctrl+e"})...),
			key.WithHelp("ctrl+e", "edit"),
		),
		StateSelector: key.NewBinding(
			key.WithKeys(getKeys("state_selector", []string{"ctrl+s"})...),
			key.WithHelp("ctrl+s", "select state"),
		),
		SessionSelector: key.NewBinding(
			key.WithKeys(getKeys("session_selector", []string{"ctrl+n"})...),
			key.WithHelp("ctrl+n", "select session"),
		),
		SetFavourite: key.NewBinding(
			key.WithKeys(getKeys("set_favourite", []string{"ctrl+f"})...),
			key.WithHelp("ctrl+f", "set favourite"),
		),
		PrevPage: key.NewBinding(
			key.WithKeys(getKeys("prev_page", []string{"h", "pgup", "b", "u"})...),
			key.WithHelp("h/pgup", "prev page"),
		),
		NextPage: key.NewBinding(
			key.WithKeys(getKeys("next_page", []string{"l", "pgdown", "f", "d"})...),
			key.WithHelp("l/pgdn", "next page"),
		),
		ForceQuit: key.NewBinding(
			key.WithKeys(getKeys("force_quit", []string{"ctrl+c"})...),
			key.WithHelp("ctrl+c", "force quit"),
		),
		ShowFullHelp: key.NewBinding(
			key.WithKeys(getKeys("show_full_help", []string{"?"})...),
			key.WithHelp("?", "more"),
		),
		CloseFullHelp: key.NewBinding(
			key.WithKeys(getKeys("close_full_help", []string{"?"})...),
			key.WithHelp("?", "close help"),
		),
		CancelWhileFiltering: key.NewBinding(
			key.WithKeys(getKeys("cancel_while_filtering", []string{"esc"})...),
			key.WithHelp("esc", "cancel"),
		),
		AcceptWhileFiltering: key.NewBinding(
			key.WithKeys(getKeys("accept_while_filtering", []string{"enter", "tab", "shift+tab", "ctrl+k", "up", "ctrl+j", "down"})...),
			key.WithHelp("enter", "apply filter"),
		),
		GoToStart: key.NewBinding(
			key.WithKeys(getKeys("go_to_start", []string{"home", "g"})...),
			key.WithHelp("g/home", "go to start"),
		),
		GoToEnd: key.NewBinding(
			key.WithKeys(getKeys("go_to_end", []string{"end", "G"})...),
			key.WithHelp("G/end", "go to end"),
		),
		Filter: key.NewBinding(
			key.WithKeys(getKeys("filter", []string{"/"})...),
			key.WithHelp("/", "filter"),
		),
		ClearFilter: key.NewBinding(
			key.WithKeys(getKeys("clear_filter", []string{"esc"})...),
			key.WithHelp("esc", "clear filter"),
		),
		EditIngredients: key.NewBinding(
			key.WithKeys(getKeys("edit_ingredients", []string{"i"})...),
			key.WithHelp("i", "edit ingredients"),
		),
		EditInstructions: key.NewBinding(
			key.WithKeys(getKeys("edit_instructions", []string{"s"})...),
			key.WithHelp("s", "edit instructions"),
		),
		EditAdd: key.NewBinding(
			key.WithKeys(getKeys("edit_add", []string{"a"})...),
			key.WithHelp("a", "add item"),
		),
		EditEdit: key.NewBinding(
			key.WithKeys(getKeys("edit_edit", []string{"e"})...),
			key.WithHelp("e", "edit item"),
		),
		EditDelete: key.NewBinding(
			key.WithKeys(getKeys("edit_delete", []string{"d"})...),
			key.WithHelp("d", "delete item"),
		),
	}
}
