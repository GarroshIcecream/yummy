package config

import (
	"strings"

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
	ModelSelector        key.Binding
	ThemeSelector        key.Binding
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
	RecipeSelector       key.Binding
	CommandPalette       key.Binding
	SetRating            key.Binding
	CookingMode          key.Binding
	ToggleIngredients    key.Binding
	ToggleChat           key.Binding
	ToggleTimer          key.Binding
	ResetTimer           key.Binding
	ChatScrollUp         key.Binding
	ChatScrollDown       key.Binding
}

type ManagerKeyMap struct {
	ForceQuit      key.Binding
	Back           key.Binding
	StateSelector  key.Binding
	ThemeSelector  key.Binding
	RecipeSelector key.Binding
	CommandPalette key.Binding
}

type MainMenuKeyMap struct {
	CursorUp      key.Binding
	CursorDown    key.Binding
	Enter         key.Binding
	Back          key.Binding
	Quit          key.Binding
	Help          key.Binding
	StateSelector key.Binding
}

type ListKeyMap struct {
	Add                     key.Binding
	Delete                  key.Binding
	Enter                   key.Binding
	SetFavourite            key.Binding
	ListKeyMap              list.KeyMap
	AdditionalShortHelpKeys func() []key.Binding
	AdditionalFullHelpKeys  func() []key.Binding
}

type DetailKeyMap struct {
	CursorUp    key.Binding
	CursorDown  key.Binding
	Edit        key.Binding
	SetRating   key.Binding
	CookingMode key.Binding
	Back        key.Binding
	Quit        key.Binding
	Help        key.Binding
}

type CookingKeyMap struct {
	NextStep          key.Binding
	PrevStep          key.Binding
	ToggleIngredients key.Binding
	ToggleChat        key.Binding
	ToggleTimer       key.Binding
	ResetTimer        key.Binding
	ChatScrollUp      key.Binding
	ChatScrollDown    key.Binding
	Enter             key.Binding
	Back              key.Binding
	Quit              key.Binding
}

type EditKeyMap struct {
	Edit             key.Binding
	EditIngredients  key.Binding
	EditInstructions key.Binding
	EditAdd          key.Binding
	EditEdit         key.Binding
	EditDelete       key.Binding
	Back             key.Binding
	Quit             key.Binding
	Help             key.Binding
	Enter            key.Binding
}

type ChatKeyMap struct {
	NewSession      key.Binding
	SessionSelector key.Binding
	ModelSelector   key.Binding
	Enter           key.Binding
	Back            key.Binding
	Quit            key.Binding
	Help            key.Binding
}

type StateSelectorKeyMap struct {
	CursorUp      key.Binding
	CursorDown    key.Binding
	StateSelector key.Binding
	Enter         key.Binding
	Back          key.Binding
	Quit          key.Binding
	Help          key.Binding
}

type SessionSelectorKeyMap struct {
	CursorUp             key.Binding
	CursorDown           key.Binding
	Filter               key.Binding
	ClearFilter          key.Binding
	CancelWhileFiltering key.Binding
	AcceptWhileFiltering key.Binding
	GoToStart            key.Binding
	GoToEnd              key.Binding
	PrevPage             key.Binding
	NextPage             key.Binding
	ShowFullHelp         key.Binding
	CloseFullHelp        key.Binding
	Enter                key.Binding
	SessionSelector      key.Binding
	Back                 key.Binding
	Quit                 key.Binding
	Help                 key.Binding
}

type ModelSelectorKeyMap struct {
	CursorUp             key.Binding
	CursorDown           key.Binding
	Filter               key.Binding
	ClearFilter          key.Binding
	CancelWhileFiltering key.Binding
	AcceptWhileFiltering key.Binding
	GoToStart            key.Binding
	GoToEnd              key.Binding
	PrevPage             key.Binding
	NextPage             key.Binding
	ShowFullHelp         key.Binding
	CloseFullHelp        key.Binding
	Enter                key.Binding
	ModelSelector        key.Binding
	Back                 key.Binding
	Quit                 key.Binding
	Help                 key.Binding
}

type ThemeSelectorKeyMap struct {
	CursorUp             key.Binding
	CursorDown           key.Binding
	Filter               key.Binding
	ClearFilter          key.Binding
	CancelWhileFiltering key.Binding
	AcceptWhileFiltering key.Binding
	GoToStart            key.Binding
	GoToEnd              key.Binding
	PrevPage             key.Binding
	NextPage             key.Binding
	ShowFullHelp         key.Binding
	CloseFullHelp        key.Binding
	Enter                key.Binding
	ThemeSelector        key.Binding
	Back                 key.Binding
	Quit                 key.Binding
	Help                 key.Binding
}

// NewKeyMapFromConfig creates a keymap using the keymap configuration
func NewKeyMapFromConfig(keymapConfig KeymapConfig) KeyMap {
	return createKeyMap(keymapConfig)
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

func (k KeyMap) GetManagerKeyMap() ManagerKeyMap {
	return ManagerKeyMap{
		Back:           k.Back,
		ForceQuit:      k.ForceQuit,
		StateSelector:  k.StateSelector,
		ThemeSelector:  k.ThemeSelector,
		RecipeSelector: k.RecipeSelector,
		CommandPalette: k.CommandPalette,
	}
}

func (k KeyMap) GetListKeyMap() ListKeyMap {
	listKeyMap := ListKeyMap{
		Add:          k.Add,
		Delete:       k.Delete,
		Enter:        k.Enter,
		SetFavourite: k.SetFavourite,
		AdditionalShortHelpKeys: func() []key.Binding {
			return []key.Binding{k.Add, k.Delete}
		},
		AdditionalFullHelpKeys: func() []key.Binding {
			return []key.Binding{k.Add, k.Delete, k.SetFavourite}
		},
		ListKeyMap: list.KeyMap{
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
		},
	}

	return listKeyMap
}

func (k KeyMap) GetMainMenuKeyMap() MainMenuKeyMap {
	return MainMenuKeyMap{
		CursorUp:      k.CursorUp,
		CursorDown:    k.CursorDown,
		Enter:         k.Enter,
		Back:          k.Back,
		Quit:          k.Quit,
		Help:          k.Help,
		StateSelector: k.StateSelector,
	}
}

func (k KeyMap) GetDetailKeyMap() DetailKeyMap {
	return DetailKeyMap{
		CursorUp:    k.CursorUp,
		CursorDown:  k.CursorDown,
		Edit:        k.Edit,
		SetRating:   k.SetRating,
		CookingMode: k.CookingMode,
		Back:        k.Back,
		Quit:        k.Quit,
		Help:        k.Help,
	}
}

func (k KeyMap) GetCookingKeyMap() CookingKeyMap {
	return CookingKeyMap{
		NextStep:          k.NextPage,
		PrevStep:          k.PrevPage,
		ToggleIngredients: k.ToggleIngredients,
		ToggleChat:        k.ToggleChat,
		ToggleTimer:       k.ToggleTimer,
		ResetTimer:        k.ResetTimer,
		ChatScrollUp:      k.ChatScrollUp,
		ChatScrollDown:    k.ChatScrollDown,
		Enter:             k.Enter,
		Back:              k.Back,
		Quit:              k.Quit,
	}
}

func (k KeyMap) GetEditKeyMap() EditKeyMap {
	return EditKeyMap{
		Edit:             k.Edit,
		EditIngredients:  k.EditIngredients,
		EditInstructions: k.EditInstructions,
		EditAdd:          k.EditAdd,
		EditEdit:         k.EditEdit,
		EditDelete:       k.EditDelete,
		Back:             k.Back,
		Quit:             k.Quit,
		Help:             k.Help,
		Enter:            k.Enter,
	}
}

func (k KeyMap) GetChatKeyMap() ChatKeyMap {
	return ChatKeyMap{
		NewSession:      k.NewSession,
		SessionSelector: k.SessionSelector,
		ModelSelector:   k.ModelSelector,
		Enter:           k.Enter,
		Back:            k.Back,
		Quit:            k.Quit,
		Help:            k.Help,
	}
}

func (k KeyMap) GetStateSelectorKeyMap() StateSelectorKeyMap {
	return StateSelectorKeyMap{
		CursorUp:      k.CursorUp,
		CursorDown:    k.CursorDown,
		StateSelector: k.StateSelector,
		Enter:         k.Enter,
		Back:          k.Back,
		Quit:          k.Quit,
		Help:          k.Help,
	}
}

func (k KeyMap) GetSessionSelectorKeyMap() SessionSelectorKeyMap {
	return SessionSelectorKeyMap{
		CursorUp:             k.CursorUp,
		CursorDown:           k.CursorDown,
		Filter:               k.Filter,
		ClearFilter:          k.ClearFilter,
		CancelWhileFiltering: k.CancelWhileFiltering,
		AcceptWhileFiltering: k.AcceptWhileFiltering,
		GoToStart:            k.GoToStart,
		GoToEnd:              k.GoToEnd,
		PrevPage:             k.PrevPage,
		NextPage:             k.NextPage,
		ShowFullHelp:         k.ShowFullHelp,
		CloseFullHelp:        k.CloseFullHelp,
		Enter:                k.Enter,
		SessionSelector:      k.SessionSelector,
		Back:                 k.Back,
		Quit:                 k.Quit,
		Help:                 k.Help,
	}
}

func (k KeyMap) GetModelSelectorKeyMap() ModelSelectorKeyMap {
	return ModelSelectorKeyMap{
		CursorUp:             k.CursorUp,
		CursorDown:           k.CursorDown,
		Filter:               k.Filter,
		ClearFilter:          k.ClearFilter,
		CancelWhileFiltering: k.CancelWhileFiltering,
		AcceptWhileFiltering: k.AcceptWhileFiltering,
		GoToStart:            k.GoToStart,
		GoToEnd:              k.GoToEnd,
		PrevPage:             k.PrevPage,
		NextPage:             k.NextPage,
		ShowFullHelp:         k.ShowFullHelp,
		CloseFullHelp:        k.CloseFullHelp,
		Enter:                k.Enter,
		ModelSelector:        k.ModelSelector,
		Back:                 k.Back,
		Quit:                 k.Quit,
		Help:                 k.Help,
	}
}

func (k KeyMap) GetThemeSelectorKeyMap() ThemeSelectorKeyMap {
	return ThemeSelectorKeyMap{
		CursorUp:             k.CursorUp,
		CursorDown:           k.CursorDown,
		Filter:               k.Filter,
		ClearFilter:          k.ClearFilter,
		CancelWhileFiltering: k.CancelWhileFiltering,
		AcceptWhileFiltering: k.AcceptWhileFiltering,
		GoToStart:            k.GoToStart,
		GoToEnd:              k.GoToEnd,
		PrevPage:             k.PrevPage,
		NextPage:             k.NextPage,
		ShowFullHelp:         k.ShowFullHelp,
		CloseFullHelp:        k.CloseFullHelp,
		Enter:                k.Enter,
		ThemeSelector:        k.ThemeSelector,
		Back:                 k.Back,
		Quit:                 k.Quit,
		Help:                 k.Help,
	}
}

// keyHelpDisplay returns the string to show in help for a key binding. Uses "space" for " " so it's visible.
func keyHelpDisplay(keys []string) string {
	s := strings.Join(keys, "/")
	if s == " " {
		return "space"
	}
	return s
}

// createKeyMap creates a keymap with the given custom bindings
func createKeyMap(keymapConfig KeymapConfig) KeyMap {
	return KeyMap{
		CursorUp: key.NewBinding(
			key.WithKeys(keymapConfig.CursorUp...),
			key.WithHelp(strings.Join(keymapConfig.CursorUp, "/"), "move up"),
		),
		CursorDown: key.NewBinding(
			key.WithKeys(keymapConfig.CursorDown...),
			key.WithHelp(strings.Join(keymapConfig.CursorDown, "/"), "move down"),
		),
		Yes: key.NewBinding(
			key.WithKeys(keymapConfig.Yes...),
			key.WithHelp(strings.Join(keymapConfig.Yes, "/"), "yes"),
		),
		No: key.NewBinding(
			key.WithKeys(keymapConfig.No...),
			key.WithHelp(strings.Join(keymapConfig.No, "/"), "no"),
		),
		Add: key.NewBinding(
			key.WithKeys(keymapConfig.Add...),
			key.WithHelp(strings.Join(keymapConfig.Add, "/"), "add recipe"),
		),
		NewSession: key.NewBinding(
			key.WithKeys(keymapConfig.NewSession...),
			key.WithHelp(strings.Join(keymapConfig.NewSession, "/"), "new session"),
		),
		Back: key.NewBinding(
			key.WithKeys(keymapConfig.Back...),
			key.WithHelp(strings.Join(keymapConfig.Back, "/"), "go back"),
		),
		Delete: key.NewBinding(
			key.WithKeys(keymapConfig.Delete...),
			key.WithHelp(strings.Join(keymapConfig.Delete, "/"), "delete recipe"),
		),
		Enter: key.NewBinding(
			key.WithKeys(keymapConfig.Enter...),
			key.WithHelp(strings.Join(keymapConfig.Enter, "/"), "select"),
		),
		Help: key.NewBinding(
			key.WithKeys(keymapConfig.Help...),
			key.WithHelp(strings.Join(keymapConfig.Help, "/"), "help"),
		),
		Edit: key.NewBinding(
			key.WithKeys(keymapConfig.Edit...),
			key.WithHelp(strings.Join(keymapConfig.Edit, "/"), "edit"),
		),
		StateSelector: key.NewBinding(
			key.WithKeys(keymapConfig.StateSelector...),
			key.WithHelp(strings.Join(keymapConfig.StateSelector, "/"), "select state"),
		),
		SessionSelector: key.NewBinding(
			key.WithKeys(keymapConfig.SessionSelector...),
			key.WithHelp(strings.Join(keymapConfig.SessionSelector, "/"), "select session"),
		),
		ModelSelector: key.NewBinding(
			key.WithKeys(keymapConfig.ModelSelector...),
			key.WithHelp(strings.Join(keymapConfig.ModelSelector, "/"), "select model"),
		),
		ThemeSelector: key.NewBinding(
			key.WithKeys(keymapConfig.ThemeSelector...),
			key.WithHelp(strings.Join(keymapConfig.ThemeSelector, "/"), "select theme"),
		),
		SetFavourite: key.NewBinding(
			key.WithKeys(keymapConfig.SetFavourite...),
			key.WithHelp(strings.Join(keymapConfig.SetFavourite, "/"), "set favourite"),
		),
		PrevPage: key.NewBinding(
			key.WithKeys(keymapConfig.PrevPage...),
			key.WithHelp(strings.Join(keymapConfig.PrevPage, "/"), "prev page"),
		),
		NextPage: key.NewBinding(
			key.WithKeys(keymapConfig.NextPage...),
			key.WithHelp(strings.Join(keymapConfig.NextPage, "/"), "next page"),
		),
		Quit: key.NewBinding(
			key.WithKeys(keymapConfig.Quit...),
			key.WithHelp(strings.Join(keymapConfig.Quit, "/"), "quit"),
		),
		ForceQuit: key.NewBinding(
			key.WithKeys(keymapConfig.ForceQuit...),
			key.WithHelp(strings.Join(keymapConfig.ForceQuit, "/"), "force quit"),
		),
		ShowFullHelp: key.NewBinding(
			key.WithKeys(keymapConfig.ShowFullHelp...),
			key.WithHelp(strings.Join(keymapConfig.ShowFullHelp, "/"), "more"),
		),
		CloseFullHelp: key.NewBinding(
			key.WithKeys(keymapConfig.CloseFullHelp...),
			key.WithHelp(strings.Join(keymapConfig.CloseFullHelp, "/"), "close help"),
		),
		CancelWhileFiltering: key.NewBinding(
			key.WithKeys(keymapConfig.CancelWhileFiltering...),
			key.WithHelp(strings.Join(keymapConfig.CancelWhileFiltering, "/"), "cancel"),
		),
		AcceptWhileFiltering: key.NewBinding(
			key.WithKeys(keymapConfig.AcceptWhileFiltering...),
			key.WithHelp(strings.Join(keymapConfig.AcceptWhileFiltering, "/"), "apply filter"),
		),
		GoToStart: key.NewBinding(
			key.WithKeys(keymapConfig.GoToStart...),
			key.WithHelp(strings.Join(keymapConfig.GoToStart, "/"), "go to start"),
		),
		GoToEnd: key.NewBinding(
			key.WithKeys(keymapConfig.GoToEnd...),
			key.WithHelp(strings.Join(keymapConfig.GoToEnd, "/"), "go to end"),
		),
		Filter: key.NewBinding(
			key.WithKeys(keymapConfig.Filter...),
			key.WithHelp(strings.Join(keymapConfig.Filter, "/"), "filter"),
		),
		ClearFilter: key.NewBinding(
			key.WithKeys(keymapConfig.ClearFilter...),
			key.WithHelp(strings.Join(keymapConfig.ClearFilter, "/"), "clear filter"),
		),
		EditIngredients: key.NewBinding(
			key.WithKeys(keymapConfig.EditIngredients...),
			key.WithHelp(strings.Join(keymapConfig.EditIngredients, "/"), "edit ingredients"),
		),
		EditInstructions: key.NewBinding(
			key.WithKeys(keymapConfig.EditInstructions...),
			key.WithHelp(strings.Join(keymapConfig.EditInstructions, "/"), "edit instructions"),
		),
		EditAdd: key.NewBinding(
			key.WithKeys(keymapConfig.EditAdd...),
			key.WithHelp(strings.Join(keymapConfig.EditAdd, "/"), "add item"),
		),
		EditEdit: key.NewBinding(
			key.WithKeys(keymapConfig.EditEdit...),
			key.WithHelp(strings.Join(keymapConfig.EditEdit, "/"), "edit item"),
		),
		EditDelete: key.NewBinding(
			key.WithKeys(keymapConfig.EditDelete...),
			key.WithHelp(strings.Join(keymapConfig.EditDelete, "/"), "delete item"),
		),
		RecipeSelector: key.NewBinding(
			key.WithKeys(keymapConfig.RecipeSelector...),
			key.WithHelp(strings.Join(keymapConfig.RecipeSelector, "/"), "find recipe"),
		),
		CommandPalette: key.NewBinding(
			key.WithKeys(keymapConfig.CommandPalette...),
			key.WithHelp(strings.Join(keymapConfig.CommandPalette, "/"), "command palette"),
		),
		SetRating: key.NewBinding(
			key.WithKeys(keymapConfig.SetRating...),
			key.WithHelp(strings.Join(keymapConfig.SetRating, "/"), "rate recipe"),
		),
		CookingMode: key.NewBinding(
			key.WithKeys(keymapConfig.CookingMode...),
			key.WithHelp(strings.Join(keymapConfig.CookingMode, "/"), "cooking mode"),
		),
		ToggleIngredients: key.NewBinding(
			key.WithKeys(keymapConfig.ToggleIngredients...),
			key.WithHelp(strings.Join(keymapConfig.ToggleIngredients, "/"), "toggle ingredients"),
		),
		ToggleChat: key.NewBinding(
			key.WithKeys(keymapConfig.ToggleChat...),
			key.WithHelp(strings.Join(keymapConfig.ToggleChat, "/"), "ask AI"),
		),
		ToggleTimer: key.NewBinding(
			key.WithKeys(keymapConfig.ToggleTimer...),
			key.WithHelp(keyHelpDisplay(keymapConfig.ToggleTimer), "start/stop timer"),
		),
		ResetTimer: key.NewBinding(
			key.WithKeys(keymapConfig.ResetTimer...),
			key.WithHelp(strings.Join(keymapConfig.ResetTimer, "/"), "reset timer"),
		),
		ChatScrollUp: key.NewBinding(
			key.WithKeys(keymapConfig.ChatScrollUp...),
			key.WithHelp(strings.Join(keymapConfig.ChatScrollUp, "/"), "scroll chat up"),
		),
		ChatScrollDown: key.NewBinding(
			key.WithKeys(keymapConfig.ChatScrollDown...),
			key.WithHelp(strings.Join(keymapConfig.ChatScrollDown, "/"), "scroll chat down"),
		),
	}
}
