package chat

import (
	"fmt"
	"regexp"
	"strings"

	themes "github.com/GarroshIcecream/yummy/internal/themes"
)

// mentionState tracks the @-mention autocomplete popup.
type mentionState struct {
	active      bool               // whether the popup is visible
	query       string             // current search text after @
	suggestions []RecipeSuggestion // filtered results
	cursor      int                // selected index
}

// mentionRe matches @RecipeName references in submitted text. Recipe names are
// wrapped in square brackets to allow spaces: @[Chicken Tikka Masala].
var mentionRe = regexp.MustCompile(`@\[([^\]]+)\]`)

// updateMention inspects the current textarea value and decides whether to
// show/hide the autocomplete popup. Returns true if the mention state changed.
func (m *mentionState) updateMention(textareaValue string, cursorCol int, executor *ExecutorService) bool {
	// Find the last @ before the cursor position.
	// We look at the entire text up to the cursor.
	before := textareaValue
	if cursorCol >= 0 && cursorCol < len(before) {
		before = before[:cursorCol]
	}

	atIdx := strings.LastIndex(before, "@")
	if atIdx < 0 {
		if m.active {
			m.reset()
			return true
		}
		return false
	}

	// Check there is no space before the @ (or it's at position 0) — the
	// mention trigger should be at the start or after a space.
	if atIdx > 0 && before[atIdx-1] != ' ' && before[atIdx-1] != '\n' {
		if m.active {
			m.reset()
			return true
		}
		return false
	}

	query := before[atIdx+1:]

	// If there is a closing bracket already we're past the mention.
	if strings.Contains(query, "]") {
		if m.active {
			m.reset()
			return true
		}
		return false
	}

	// Strip leading [ if the user already typed it
	query = strings.TrimPrefix(query, "[")

	// Search recipes
	suggestions := executor.SearchRecipeNames(query)

	changed := !m.active || m.query != query
	m.active = len(suggestions) > 0
	m.query = query
	m.suggestions = suggestions
	if m.cursor >= len(suggestions) {
		m.cursor = max(0, len(suggestions)-1)
	}
	return changed
}

// accept returns the completed mention text that should replace the partial
// @query in the textarea, and resets the state.
func (m *mentionState) accept() (string, uint) {
	if !m.active || len(m.suggestions) == 0 {
		return "", 0
	}
	sel := m.suggestions[m.cursor]
	m.reset()
	return fmt.Sprintf("@[%s]", sel.Name), sel.ID
}

func (m *mentionState) moveUp() {
	if m.cursor > 0 {
		m.cursor--
	}
}

func (m *mentionState) moveDown() {
	if m.cursor < len(m.suggestions)-1 {
		m.cursor++
	}
}

func (m *mentionState) reset() {
	m.active = false
	m.query = ""
	m.suggestions = nil
	m.cursor = 0
}

// viewMention renders the autocomplete popup (rendered above the textarea).
func viewMention(ms *mentionState, theme *themes.Theme, width int) string {
	if !ms.active || len(ms.suggestions) == 0 {
		return ""
	}

	popupWidth := min(width-4, 50)
	selectedStyle := theme.ChatMentionPopupSelected.Width(popupWidth)
	normalStyle := theme.ChatMentionPopupItem.Width(popupWidth)

	var rows []string
	rows = append(rows, theme.ChatMentionPopupHeader.Render("Recipes"))

	for i, s := range ms.suggestions {
		label := s.Name
		if i == ms.cursor {
			rows = append(rows, selectedStyle.Render("› "+label))
		} else {
			rows = append(rows, normalStyle.Render("  "+label))
		}
	}

	return theme.ChatMentionPopupBorder.Width(popupWidth).Render(strings.Join(rows, "\n"))
}

// HighlightMentions finds all @[RecipeName] patterns in already-rendered
// content (e.g. after glamour markdown rendering) and wraps them with a
// distinct style so recipe mentions stand out in the chat view.
func HighlightMentions(renderedContent string, theme *themes.Theme) string {
	matches := mentionRe.FindAllStringIndex(renderedContent, -1)
	if len(matches) == 0 {
		return renderedContent
	}

	style := theme.ChatMention
	var result strings.Builder
	lastEnd := 0

	for _, loc := range matches {
		start, end := loc[0], loc[1]
		mention := renderedContent[start:end]
		result.WriteString(renderedContent[lastEnd:start])
		result.WriteString(style.Render(mention))
		lastEnd = end
	}
	result.WriteString(renderedContent[lastEnd:])
	return result.String()
}

// resolveMentions finds all @[RecipeName] references in the user input,
// fetches full recipe data, and returns the augmented prompt to send to the
// LLM. The display text (shown in the conversation) keeps the @[Name] markers
// but the prompt text prepended to the LLM includes the recipe content.
func resolveMentions(userInput string, executor *ExecutorService) string {
	matches := mentionRe.FindAllStringSubmatch(userInput, -1)
	if len(matches) == 0 {
		return userInput
	}

	// Collect unique recipe names
	seen := make(map[string]bool)
	var recipeContexts []string

	for _, match := range matches {
		name := match[1]
		if seen[name] {
			continue
		}
		seen[name] = true

		// Look up recipe by name
		suggestions := executor.SearchRecipeNames(name)
		for _, s := range suggestions {
			if strings.EqualFold(s.Name, name) {
				content, err := executor.GetFullRecipe(s.ID)
				if err == nil {
					recipeContexts = append(recipeContexts,
						fmt.Sprintf("--- Referenced Recipe: %s ---\n%s\n---", s.Name, content))
				}
				break
			}
		}
	}

	if len(recipeContexts) == 0 {
		return userInput
	}

	// Build the augmented prompt: recipe context first, then the user message.
	var sb strings.Builder
	sb.WriteString("The user is referencing the following recipe(s) from their cookbook:\n\n")
	for _, ctx := range recipeContexts {
		sb.WriteString(ctx)
		sb.WriteString("\n\n")
	}
	sb.WriteString("User message: ")
	sb.WriteString(userInput)
	return sb.String()
}
