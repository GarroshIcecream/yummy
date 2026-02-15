package dialog

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"strconv"
	"strings"

	"github.com/GarroshIcecream/yummy/internal/config"
	db "github.com/GarroshIcecream/yummy/internal/db"
	common "github.com/GarroshIcecream/yummy/internal/models/common"
	messages "github.com/GarroshIcecream/yummy/internal/models/msg"
	"github.com/GarroshIcecream/yummy/internal/scrape"
	themes "github.com/GarroshIcecream/yummy/internal/themes"
	"github.com/GarroshIcecream/yummy/internal/utils"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var addRecipeSubmitKey = key.NewBinding(key.WithKeys("enter"))

type AddRecipeFromURLDialogCmp struct {
	cookbook           *db.CookBook
	urlInput           textinput.Model
	spinner            spinner.Model
	width              int
	height             int
	keyMap             config.KeyMap
	theme              *themes.Theme
	loading            bool
	loadingText        string // current spinner message
	errorMsg           string
	fetchedURL         string
	pythonPath         string // optional path to Python for recipe-scrapers (from config)
	existingID         uint   // non-zero when URL already exists in cookbook
	llmIngredientModel string // Ollama model name for ingredient parsing
}

func NewAddRecipeFromURLDialog(cookbook *db.CookBook, theme *themes.Theme) (*AddRecipeFromURLDialogCmp, error) {
	cfg := config.GetGlobalConfig()
	if cfg == nil {
		return nil, fmt.Errorf("global config not set")
	}

	dialogConfig := cfg.AddRecipeFromURLDialog

	ti := textinput.New()
	ti.Placeholder = "https://example.com/recipe"
	if w := dialogConfig.Width - 16; w > 24 {
		ti.Width = w
	} else {
		ti.Width = 48
	}
	ti.Focus()

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = theme.AddRecipeFromURLSpinner

	// Resolve LLM model name: use dedicated model if set, otherwise fall back to chat default.
	llmModel := dialogConfig.LLMIngredientModel
	if llmModel == "" {
		llmModel = cfg.Chat.DefaultModel
	}

	return &AddRecipeFromURLDialogCmp{
		cookbook:           cookbook,
		urlInput:           ti,
		spinner:            s,
		keyMap:             cfg.Keymap.ToKeyMap(),
		width:              dialogConfig.Width,
		height:             dialogConfig.Height,
		theme:              theme,
		pythonPath:         dialogConfig.PythonPath,
		llmIngredientModel: llmModel,
	}, nil
}

func (m *AddRecipeFromURLDialogCmp) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, m.spinner.Tick)
}

func (m *AddRecipeFromURLDialogCmp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)

	case scrapeProgressMsg:
		m.loadingText = msg.text
		return m, nil

	case scrapeDoneMsg:
		// Scraping done — pick a fun quip for the ingredient parsing phase.
		quip := ingredientParsingQuips[rand.Intn(len(ingredientParsingQuips))]
		m.loadingText = quip
		return m, parseAndSaveCmd(msg.scraper, msg.url, msg.llmModel, m.cookbook)

	case scrapeAndSaveResultMsg:
		m.loading = false
		if msg.err != nil {
			m.errorMsg = friendlyScrapeError(msg.err)
			return m, nil
		}
		statusMsg := config.GetListConfig().ViewStatusMessageRecipeAdded
		cmds = append(cmds, messages.SendCloseModalViewMsg())
		cmds = append(cmds, messages.SendSessionStateMsg(common.SessionStateList))
		cmds = append(cmds, messages.SendRecipeAddedFromURLMsg(msg.recipeID, statusMsg))
		return m, tea.Sequence(cmds...)

	case tea.KeyMsg:
		if m.loading {
			return m, nil
		}
		if key.Matches(msg, addRecipeSubmitKey) {
			// Second Enter when a duplicate was found — navigate to that recipe
			if m.existingID != 0 {
				cmds = append(cmds, messages.SendCloseModalViewMsg())
				cmds = append(cmds, messages.SendSessionStateMsg(common.SessionStateList))
				cmds = append(cmds, messages.SendRecipeAddedFromURLMsg(m.existingID, ""))
				return m, tea.Sequence(cmds...)
			}

			urlStr := m.urlInput.Value()
			if urlStr == "" {
				m.errorMsg = "Please enter a URL"
				return m, nil
			}
			if err := utils.ValidateURL(urlStr); err != nil {
				m.errorMsg = err.Error()
				return m, nil
			}

			// Check for duplicate URL in the cookbook
			recipeID, err := m.cookbook.RecipeExistsByURL(urlStr)
			if err != nil {
				slog.Error("Failed to check for duplicate URL", "error", err)
			} else if recipeID != 0 {
				m.existingID = recipeID
				return m, nil
			}

			m.errorMsg = ""
			m.fetchedURL = urlStr
			m.loading = true

			cmds = append(cmds, m.spinner.Tick, scrapeURLCmd(
				urlStr, m.pythonPath, m.llmIngredientModel,
			))
		}
	}

	prevValue := m.urlInput.Value()
	var cmd tea.Cmd
	m.urlInput, cmd = m.urlInput.Update(msg)
	cmds = append(cmds, cmd)
	// Reset duplicate state when the user edits the URL
	if m.urlInput.Value() != prevValue {
		m.existingID = 0
	}
	return m, tea.Batch(cmds...)
}

func (m *AddRecipeFromURLDialogCmp) View() string {
	innerWidth := m.width - 6 // account for dialog border (2) + padding (4)
	if innerWidth < 40 {
		innerWidth = 40
	}

	// Thick block border for a floating-panel / card look
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

	// All elements share the same width so nothing drifts.
	row := func(s lipgloss.Style, align lipgloss.Position, text string) string {
		return s.Width(innerWidth).Align(align).Render(text)
	}

	// Title — bold, centered
	title := row(m.theme.AddRecipeFromURLTitle, lipgloss.Center, "Add Recipe from URL")

	// Separator
	sep := m.theme.AddRecipeFromURLSeparator.Render(strings.Repeat("─", innerWidth))

	// Label
	label := row(m.theme.AddRecipeFromURLPrompt, lipgloss.Center, "Paste a recipe URL")

	// Input field — material-design inspired: bottom line only, no box
	inputBorder := lipgloss.Border{Bottom: "━"}
	inputBox := lipgloss.NewStyle().
		BorderStyle(inputBorder).
		BorderBottom(true).
		BorderForeground(m.theme.AddRecipeFromURLAccent.GetForeground()).
		Width(innerWidth).
		Padding(0, 1).
		Render(m.urlInput.View())

	// Status
	var status string
	if m.existingID != 0 {
		status = row(m.theme.AddRecipeFromURLPrompt, lipgloss.Center,
			"Recipe already in cookbook — press enter to view it")
	} else if m.loading {
		spinnerText := m.loadingText
		if spinnerText == "" {
			spinnerText = "Fetching recipe…"
		}
		status = lipgloss.NewStyle().Width(innerWidth).Align(lipgloss.Center).Render(
			m.theme.AddRecipeFromURLSpinner.Render(m.spinner.View()) +
				m.theme.AddRecipeFromURLPrompt.Render(" "+spinnerText))
	} else if m.errorMsg != "" {
		status = row(m.theme.AddRecipeFromURLError, lipgloss.Center, m.errorMsg)
	}

	// Help — subtle keys, centered
	enterHelp := m.keyMap.Enter.Help().Key
	escHelp := m.keyMap.Quit.Help().Key
	enterKey := m.theme.AddRecipeFromURLKeyHighlight.Render(enterHelp)
	escKey := m.theme.AddRecipeFromURLKeyHighlight.Render(escHelp)
	dot := m.theme.AddRecipeFromURLHelp.Render(" · ")
	helpText := enterKey +
		m.theme.AddRecipeFromURLHelp.Render(" submit") + dot +
		escKey + m.theme.AddRecipeFromURLHelp.Render(" cancel")
	helpLine := lipgloss.NewStyle().Width(innerWidth).Align(lipgloss.Center).Render(helpText)

	// Assemble — all rows are innerWidth so left-join keeps them flush
	parts := []string{title, sep, "", label, inputBox}
	if status != "" {
		parts = append(parts, "", status)
	}
	parts = append(parts, "", helpLine)

	content := lipgloss.JoinVertical(lipgloss.Left, parts...)

	// No Height set — dialog shrinks to fit content
	rendered := lipgloss.NewStyle().
		Border(dialogBorder).
		BorderForeground(m.theme.AddRecipeFromURLAccent.GetForeground()).
		Padding(1, 2).
		Width(m.width).
		Render(content)

	return m.theme.AddRecipeFromURLContainer.Render(rendered)
}

func (m *AddRecipeFromURLDialogCmp) SetSize(width, height int) {
	if width > 0 {
		m.width = width - 20
		if m.width < 50 {
			m.width = 50
		}
		if m.width > 70 {
			m.width = 70
		}
	}
	m.height = height
	// textinput.Width controls how many chars are visible before horizontal
	// scrolling kicks in. Account for: dialog border (2) + dialog padding (4)
	// + input padding (2) = 8 chars of overhead.
	if w := m.width - 8; w > 20 {
		m.urlInput.Width = w
	} else if m.urlInput.Width <= 0 {
		m.urlInput.Width = 48
	}
}

func (m *AddRecipeFromURLDialogCmp) GetSize() (int, int) {
	return m.width, m.height
}

func (m *AddRecipeFromURLDialogCmp) GetModelState() common.ModelState {
	return common.ModelStateLoaded
}

// scrapeProgressMsg updates the spinner text during the scrape-and-save pipeline.
type scrapeProgressMsg struct {
	text string
}

// scrapeDoneMsg carries the scraped data to the next phase (ingredient parsing + save).
type scrapeDoneMsg struct {
	scraper  scrape.Scraper
	url      string
	llmModel string
}

// scrapeAndSaveResultMsg is the internal result of the full pipeline.
type scrapeAndSaveResultMsg struct {
	recipeID uint
	err      error
}

// ingredientParsingQuips are shown while the LLM is crunching ingredients.
var ingredientParsingQuips = []string{
	"Decoding the grocery list…",
	"Teaching AI to read a recipe…",
	"Sorting the spice rack…",
	"Chopping ingredients into data…",
	"Mise en place in progress…",
	"Convincing the AI that tsp ≠ tbsp…",
	"Translating chef-speak…",
}

// scrapeURLCmd handles only the URL scraping phase.
func scrapeURLCmd(url, pythonPath, llmModel string) tea.Cmd {
	return func() tea.Msg {
		scraper, err := scrape.ScrapeURL(url, pythonPath)
		if err != nil {
			return scrapeAndSaveResultMsg{err: err}
		}
		return scrapeDoneMsg{scraper: scraper, url: url, llmModel: llmModel}
	}
}

// parseAndSaveCmd handles ingredient parsing (with LLM) and saving to DB.
func parseAndSaveCmd(s scrape.Scraper, url, llmModel string, cookbook *db.CookBook) tea.Cmd {
	return func() tea.Msg {
		recipeData := recipeRawFromScraper(s, url, llmModel)

		recipeID, err := cookbook.SaveScrapedRecipe(recipeData)
		if err != nil {
			slog.Error("Failed to save recipe from URL", "error", err)
			return scrapeAndSaveResultMsg{err: fmt.Errorf("save error: %w", err)}
		}

		return scrapeAndSaveResultMsg{recipeID: recipeID}
	}
}

// friendlyScrapeError turns a raw scrape error into a short, user-friendly message.
func friendlyScrapeError(err error) string {
	msg := err.Error()
	low := strings.ToLower(msg)

	switch {
	case strings.Contains(low, "isn't currently supported") ||
		strings.Contains(low, "not supported"):
		return "Can't scrape this website yet"
	case strings.Contains(low, "no python found"):
		return "Python 3 is needed for scraping"
	case strings.Contains(low, "no such host") ||
		strings.Contains(low, "dial tcp") ||
		strings.Contains(low, "connection refused"):
		return "Can't reach that website"
	case strings.Contains(low, "404") || strings.Contains(low, "not found"):
		return "Page not found"
	case strings.Contains(low, "timeout") || strings.Contains(low, "deadline exceeded"):
		return "Request timed out"
	case strings.Contains(low, "no recipe found") || strings.Contains(low, "no schema found"):
		return "No recipe found on this page"
	default:
		if len(msg) > 80 {
			msg = msg[:77] + "..."
		}
		return msg
	}
}

// recipeRawFromScraper converts a scrape.Scraper into utils.RecipeRaw for saving.
// When useLLM is true it attempts to parse ingredients via Ollama (experimental),
// falling back to the regex parser on any error.
func recipeRawFromScraper(s scrape.Scraper, sourceURL string, llmModel string) *utils.RecipeRaw {
	r := &utils.RecipeRaw{
		Metadata: utils.RecipeMetadata{
			URL:          sourceURL,
			Ingredients:  []utils.Ingredient{},
			Instructions: []string{},
			Categories:   []string{},
		},
	}

	if name, ok := s.Name(); ok && name != "" {
		r.RecipeName = strings.TrimSpace(name)
	} else {
		r.RecipeName = "Imported Recipe #" + strconv.Itoa(rand.Intn(1000))
	}
	if desc, ok := s.Description(); ok {
		r.RecipeDescription = strings.TrimSpace(desc)
	}
	if author, ok := s.Author(); ok {
		r.Metadata.Author = strings.TrimSpace(author)
	}
	if ct, ok := s.CookTime(); ok {
		r.Metadata.CookTime = ct
	}
	if pt, ok := s.PrepTime(); ok {
		r.Metadata.PrepTime = pt
	}
	if tt, ok := s.TotalTime(); ok {
		r.Metadata.TotalTime = tt
	}
	if y, ok := s.Yields(); ok && y != "" {
		r.Metadata.Quantity = strings.TrimSpace(y)
	}
	if ingList, ok := s.Ingredients(); ok {
		r.Metadata.Ingredients = parseIngredientList(ingList, llmModel)
	}
	if instr, ok := s.Instructions(); ok {
		for _, step := range instr {
			if t := strings.TrimSpace(step); t != "" {
				r.Metadata.Instructions = append(r.Metadata.Instructions, t)
			}
		}
	}
	if cat, ok := s.Categories(); ok {
		for _, c := range cat {
			if t := strings.TrimSpace(c); t != "" {
				r.Metadata.Categories = append(r.Metadata.Categories, t)
			}
		}
	}
	return r
}

// parseIngredientList parses a list of raw ingredient strings. When useLLM is
// true it sends the entire list to an Ollama model for structured extraction
// and falls back to the regex parser if the LLM call fails. Even on the regex
// fallback path, LLM is still used to extract base names for highlighting.
func parseIngredientList(rawList []string, llmModel string) []utils.Ingredient {
	// Clean up the raw list first.
	cleaned := make([]string, 0, len(rawList))
	for _, s := range rawList {
		if t := strings.TrimSpace(s); t != "" {
			cleaned = append(cleaned, t)
		}
	}
	if len(cleaned) == 0 {
		return []utils.Ingredient{}
	}

	// Experimental: try full LLM-based parsing (amount + unit + name + base_name).
	if llmModel != "" {
		slog.Info("Attempting LLM ingredient parsing", "model", llmModel, "count", len(cleaned))
		parsed, err := utils.ParseIngredientsWithLLM(context.Background(), cleaned, llmModel)
		if err != nil {
			slog.Error("LLM ingredient parsing failed, falling back to regex", "error", err)
		} else if len(parsed) > 0 {
			slog.Info("LLM ingredient parsing succeeded", "count", len(parsed))
			return parsed
		}
	}

	// Regex fallback for amount/unit/name parsing.
	ingredients := make([]utils.Ingredient, 0, len(cleaned))
	for _, ingStr := range cleaned {
		if parsed, err := utils.ParseIngredient(ingStr); err == nil {
			ingredients = append(ingredients, parsed)
		} else {
			ingredients = append(ingredients, utils.Ingredient{Name: ingStr})
		}
	}

	// Even on the regex path, try the lightweight LLM base-name extraction
	// so we get good highlighting tokens.
	if llmModel != "" && len(ingredients) > 0 {
		slog.Info("Attempting LLM base name extraction", "model", llmModel, "count", len(ingredients))
		if err := utils.ExtractBaseNamesWithLLM(context.Background(), ingredients, llmModel); err != nil {
			slog.Error("LLM base name extraction failed, using full names for highlighting", "error", err)
		}
	}

	return ingredients
}
