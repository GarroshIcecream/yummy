package dialog

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/GarroshIcecream/yummy/internal/config"
	db "github.com/GarroshIcecream/yummy/internal/db"
	yllms "github.com/GarroshIcecream/yummy/internal/llms"
	"github.com/GarroshIcecream/yummy/internal/llms/chatgpt"
	common "github.com/GarroshIcecream/yummy/internal/models/common"
	messages "github.com/GarroshIcecream/yummy/internal/models/msg"
	"github.com/GarroshIcecream/yummy/internal/scrape"
	themes "github.com/GarroshIcecream/yummy/internal/themes"
	"github.com/GarroshIcecream/yummy/internal/utils"
	langllms "github.com/tmc/langchaingo/llms"
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
	llmIngredientModel string // model for ingredient parsing ("" = provider default)
	aiParsing          bool   // whether any AI step of the import runs
	aiFallback         bool   // ChatGPT is active, so pages recipe-scrapers can't read can be read by AI
	aiOffer            bool   // scraping failed and enter now imports the page with AI
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
		ti.SetWidth(w)
	} else {
		ti.SetWidth(48)
	}
	ti.Focus()

	s := spinner.New()
	s.Spinner = spinner.Dot

	return &AddRecipeFromURLDialogCmp{
		cookbook:           cookbook,
		urlInput:           ti,
		spinner:            s,
		keyMap:             cfg.Keymap.ToKeyMap(),
		width:              dialogConfig.Width,
		height:             dialogConfig.Height,
		theme:              theme,
		pythonPath:         dialogConfig.PythonPath,
		llmIngredientModel: dialogConfig.LLMIngredientModel,
		aiParsing:          dialogConfig.AIParsing,
		aiFallback:         cfg.Chat.Provider == "chatgpt",
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
		return m, parseAndSaveCmd(msg.scraper, msg.url, msg.llmModel, msg.aiParsing, m.cookbook)

	case scrapeFailedMsg:
		m.loading = false
		m.errorMsg = friendlyScrapeError(msg.err)
		m.aiOffer = m.aiFallback && aiCanRead(msg.err)
		return m, nil

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

	case tea.KeyPressMsg:
		if m.loading {
			return m, nil
		}
		// tab: skip the import and write the recipe by hand instead.
		if msg.String() == "tab" {
			return m, tea.Sequence(
				messages.SendCloseModalViewMsg(),
				messages.SendSessionStateMsg(common.SessionStateEdit),
				messages.SendEditRecipeMsg(nil),
			)
		}
		if key.Matches(msg, addRecipeSubmitKey) {
			// Second Enter when a duplicate was found — navigate to that recipe
			if m.existingID != 0 {
				cmds = append(cmds, messages.SendCloseModalViewMsg())
				cmds = append(cmds, messages.SendSessionStateMsg(common.SessionStateList))
				cmds = append(cmds, messages.SendRecipeAddedFromURLMsg(m.existingID, ""))
				return m, tea.Sequence(cmds...)
			}

			if m.aiOffer {
				m.aiOffer = false
				m.errorMsg = ""
				m.loading = true
				m.loadingText = "ChatGPT is reading the page…"
				return m, tea.Batch(m.spinner.Tick,
					aiScrapeCmd(m.fetchedURL, m.llmIngredientModel, m.aiParsing))
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
				urlStr, m.pythonPath, m.llmIngredientModel, m.aiParsing,
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
		m.aiOffer = false
	}
	return m, tea.Batch(cmds...)
}

// Busy reports whether a recipe import is running.
func (m *AddRecipeFromURLDialogCmp) Busy() (bool, string) {
	return m.loading, "Importing recipe"
}

func (m *AddRecipeFromURLDialogCmp) View() tea.View {
	innerWidth := max(
		// account for dialog border (2) + padding (4)
		m.width-6, 40)

	// All elements share the same width so nothing drifts.
	row := func(s lipgloss.Style, align lipgloss.Position, text string) string {
		return s.Width(innerWidth).Align(align).Render(text)
	}

	// Title — bold, centered
	title := row(m.theme.AddRecipeFromURLTitle, lipgloss.Center, "Add Recipe")

	// Separator
	sep := m.theme.AddRecipeFromURLSeparator.Render(strings.Repeat("─", innerWidth))

	// Label
	label := row(m.theme.AddRecipeFromURLPrompt, lipgloss.Center, "Paste a recipe URL to import it")
	aiMode := "AI parsing on"
	if !m.aiParsing {
		aiMode = "AI parsing off"
	}
	aiHint := "change it in the command palette"
	if k := m.keyMap.CommandPalette.Help().Key; k != "" {
		aiHint = "change it with " + k
	}
	aiLine := row(m.theme.AddRecipeFromURLHelp, lipgloss.Center, aiMode+" · "+aiHint)

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
		if m.aiOffer {
			// A muted reason, then the way forward.
			status = row(m.theme.AddRecipeFromURLHelp, lipgloss.Center, m.errorMsg) + "\n" +
				row(m.theme.AddRecipeFromURLPrompt, lipgloss.Center,
					"Press enter to let ChatGPT read the page instead")
		} else {
			status = row(m.theme.AddRecipeFromURLError, lipgloss.Center, m.errorMsg)
		}
	}

	// Help — subtle keys, centered
	enterKey := m.keyMap.Enter.Help().Key
	if enterKey == "" {
		enterKey = "enter"
	}
	enterLabel := "import"
	if m.aiOffer {
		enterLabel = "use ChatGPT"
	}
	helpItem := func(k, label string) string {
		return m.theme.AddRecipeFromURLKeyHighlight.Render(k) + m.theme.AddRecipeFromURLHelp.Render(" "+label)
	}
	items := []string{
		helpItem(enterKey, enterLabel),
		helpItem("tab", "write manually"),
		helpItem("esc", "cancel"),
	}
	if m.loading {
		items = items[2:] // enter and tab do nothing mid-import
	}
	helpLine := packHelp(items, m.theme.AddRecipeFromURLHelp.Render(" · "), innerWidth)

	// Assemble — all rows are innerWidth so left-join keeps them flush
	parts := []string{title, sep, "", label, inputBox}
	if status != "" {
		parts = append(parts, "", status)
	}
	parts = append(parts, "", aiLine, helpLine)

	content := lipgloss.JoinVertical(lipgloss.Left, parts...)
	// Fill the interior with the elevated surface so the dialog reads as a solid
	// card, not a bordered box with a transparent interior showing the canvas
	// behind it (the foreground-only rows leave bare cells otherwise).
	content = utils.FillBackground(content, m.theme.SurfaceBg, innerWidth)

	// No Height set — dialog shrinks to fit content
	rendered := m.theme.AddRecipeFromURLDialog.
		Width(m.width).
		Render(content)

	return tea.NewView(m.theme.AddRecipeFromURLContainer.Render(rendered))
}

// packHelp lays out help items on centered lines no wider than width,
// breaking only between items so a key never ends up apart from its label.
func packHelp(items []string, sep string, width int) string {
	var lines []string
	cur := ""
	for _, it := range items {
		if cur != "" && lipgloss.Width(cur+sep+it) > width {
			lines = append(lines, cur)
			cur = ""
		}
		if cur == "" {
			cur = it
		} else {
			cur += sep + it
		}
	}
	if cur != "" {
		lines = append(lines, cur)
	}
	center := lipgloss.NewStyle().Width(width).Align(lipgloss.Center)
	for i, l := range lines {
		lines[i] = center.Render(l)
	}
	return strings.Join(lines, "\n")
}

func (m *AddRecipeFromURLDialogCmp) SetSize(width, height int) {
	if width > 0 {
		m.width = min(clampModalWidth(width, m.width, 40), 70)
	}
	m.height = clampModalHeight(height, m.height, 10)
	// textinput.Width controls how many chars are visible before horizontal
	// scrolling kicks in. Account for: dialog border (2) + dialog padding (4)
	// + input padding (2) + the "> " prompt (2) + the cursor cell (1) = 11.
	// Any less and a long URL wraps onto a second line instead of scrolling.
	if w := m.width - 11; w > 20 {
		m.urlInput.SetWidth(w)
	} else if m.urlInput.Width() <= 0 {
		m.urlInput.SetWidth(48)
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
	scraper   scrape.Scraper
	url       string
	llmModel  string
	aiParsing bool
}

// scrapeFailedMsg reports that recipe-scrapers couldn't read the page.
type scrapeFailedMsg struct {
	err error
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
func scrapeURLCmd(url, pythonPath, llmModel string, aiParsing bool) tea.Cmd {
	return func() tea.Msg {
		scraper, err := scrape.ScrapeURL(url, pythonPath)
		if err != nil {
			return scrapeFailedMsg{err: err}
		}
		return scrapeDoneMsg{scraper: scraper, url: url, llmModel: llmModel, aiParsing: aiParsing}
	}
}

// aiScrapeTimeout bounds downloading the page and the model reading it.
const aiScrapeTimeout = 3 * time.Minute

// aiScrapeCmd downloads the page and has the active model (ChatGPT) read the
// recipe out of it, for sites recipe-scrapers doesn't support. The result
// continues through the same parse-and-save phase as a scraped recipe.
func aiScrapeCmd(url, llmModel string, aiParsing bool) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), aiScrapeTimeout)
		defer cancel()
		llm, err := yllms.NewActiveModel(ctx)
		if err != nil {
			return scrapeAndSaveResultMsg{err: err}
		}
		var s scrape.Scraper
		text, err := scrape.FetchPageText(ctx, nil, url)
		switch {
		case errors.Is(err, scrape.ErrSiteBlocked):
			// The site refuses our download (e.g. a Cloudflare check), so let
			// ChatGPT open the page from OpenAI's side with its web tool.
			slog.Info("Page blocks direct download, asking ChatGPT to open it", "url", url)
			s, err = scrape.ExtractRecipeFromURLWithLLM(ctx, llm, url, chatgpt.WithWebSearch())
		case err != nil:
			return scrapeAndSaveResultMsg{err: err}
		default:
			s, err = scrape.ExtractRecipeWithLLM(ctx, llm, url, text)
		}
		if err != nil {
			slog.Error("AI recipe import failed", "url", url, "error", err)
			return scrapeAndSaveResultMsg{err: err}
		}
		return scrapeDoneMsg{scraper: s, url: url, llmModel: llmModel, aiParsing: aiParsing}
	}
}

// aiCanRead reports whether reading the page with AI might succeed after
// recipe-scrapers failed with err: not when the page itself can't be fetched.
func aiCanRead(err error) bool {
	low := strings.ToLower(err.Error())
	for _, unreachable := range []string{"no such host", "dial tcp", "connection refused", "404"} {
		if strings.Contains(low, unreachable) {
			return false
		}
	}
	return true
}

// parseAndSaveCmd handles ingredient parsing (with LLM) and saving to DB.
// With aiParsing off, or when the model can't be built, every AI step is
// skipped and the rule-based parser is used.
func parseAndSaveCmd(s scrape.Scraper, url, llmModel string, aiParsing bool, cookbook *db.CookBook) tea.Cmd {
	return func() tea.Msg {
		var llm langllms.Model
		if aiParsing {
			m, err := yllms.NewIngredientModel(llmModel)
			if err != nil {
				slog.Error("Failed to create LLM for import, using rule-based parsing", "error", err)
			} else {
				llm = m
			}
		}
		recipeData := recipeRawFromScraper(s, url, llm)
		suggestCategories(recipeData, s, llm, cookbook)

		recipeID, err := cookbook.SaveScrapedRecipe(recipeData)
		if err != nil {
			slog.Error("Failed to save recipe from URL", "error", err)
			return scrapeAndSaveResultMsg{err: fmt.Errorf("save error: %w", err)}
		}

		return scrapeAndSaveResultMsg{recipeID: recipeID}
	}
}

// friendlyScrapeError turns a raw scrape error into a short, user-friendly
// message. Raw scraper output (exit codes, Python tracebacks) is logged, never
// shown.
func friendlyScrapeError(err error) string {
	msg := err.Error()
	low := strings.ToLower(msg)
	slog.Warn("Recipe import failed", "error", msg)
	if errors.Is(err, scrape.ErrSiteBlocked) {
		return "This site blocks automated downloads, so ChatGPT can't read it either"
	}

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
	case strings.Contains(low, "http error 401") || strings.Contains(low, "http error 403") || strings.Contains(low, "forbidden"):
		return "This site blocked the import"
	case strings.Contains(low, "http error 429") || strings.Contains(low, "too many requests"):
		return "This site is rate limiting imports, try again later"
	case strings.Contains(low, "http error 5") || strings.Contains(low, "service unavailable") ||
		strings.Contains(low, "bad gateway"):
		return "This site is having trouble, try again later"
	case strings.Contains(low, "timeout") || strings.Contains(low, "deadline exceeded"):
		return "Request timed out"
	case strings.Contains(low, "no recipe found") || strings.Contains(low, "no schema found"):
		return "No recipe found on this page"
	case strings.Contains(low, "exit status") || strings.Contains(low, "stderr") ||
		strings.Contains(low, "traceback"):
		return "Couldn't read a recipe from this page"
	default:
		if len(msg) > 80 {
			msg = msg[:77] + "..."
		}
		return msg
	}
}

// recipeRawFromScraper converts a scrape.Scraper into utils.RecipeRaw for saving.
// When llm is non-nil it attempts to parse ingredients via the LLM
// (experimental), falling back to the regex parser on any error.
func recipeRawFromScraper(s scrape.Scraper, sourceURL string, llm langllms.Model) *utils.RecipeRaw {
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
	if groups, ok := s.IngredientGroups(); ok {
		// Check if any groups are missing a purpose label.
		needsLabels := false
		if len(groups) > 1 {
			for _, g := range groups {
				if g.Purpose == nil || strings.TrimSpace(*g.Purpose) == "" {
					needsLabels = true
					break
				}
			}
		}

		// Ask LLM to generate labels for unlabelled groups.
		var llmLabels []string
		if needsLabels && llm != nil {
			unlabelled := make([][]string, len(groups))
			for i, g := range groups {
				unlabelled[i] = g.Ingredients
			}
			labels, err := utils.LabelIngredientGroups(context.Background(), unlabelled, llm)
			if err != nil {
				slog.Error("LLM group labelling failed, groups will be unlabelled", "error", err)
			} else {
				llmLabels = labels
			}
		}

		for i, g := range groups {
			purpose := ""
			if g.Purpose != nil {
				purpose = strings.TrimSpace(*g.Purpose)
			}
			// Fill in missing purpose from LLM labels.
			if purpose == "" && i < len(llmLabels) {
				purpose = strings.TrimSpace(llmLabels[i])
			}
			parsed := parseIngredientList(g.Ingredients, llm)
			for j := range parsed {
				parsed[j].Group = purpose
			}
			r.Metadata.Ingredients = append(r.Metadata.Ingredients, parsed...)
		}
	} else if ingList, ok := s.Ingredients(); ok {
		r.Metadata.Ingredients = parseIngredientList(ingList, llm)
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

// suggestCategories fills in categories when the site provides none: first
// from the recipe's cuisine, then (with AI parsing on) from the LLM, which is
// steered towards the cookbook's existing categories.
func suggestCategories(r *utils.RecipeRaw, s scrape.Scraper, llm langllms.Model, cookbook *db.CookBook) {
	if len(r.Metadata.Categories) > 0 {
		return
	}
	if cuisines, ok := s.Cuisine(); ok {
		r.Metadata.Categories = append(r.Metadata.Categories, cuisines...)
		return
	}
	if llm == nil {
		return
	}
	existing, err := cookbook.GetAllCategories()
	if err != nil {
		slog.Error("Failed to load categories for suggestion", "error", err)
	}
	names := make([]string, 0, len(r.Metadata.Ingredients))
	for _, ing := range r.Metadata.Ingredients {
		names = append(names, ing.Name)
	}
	cats, err := utils.SuggestCategories(context.Background(), r.RecipeName, names, existing, llm)
	if err != nil {
		slog.Error("LLM category suggestion failed, recipe stays uncategorised", "error", err)
		return
	}
	r.Metadata.Categories = cats
}

// parseIngredientList parses a list of raw ingredient strings. When llm is
// non-nil it sends the entire list to the LLM for structured extraction
// and falls back to the regex parser if the LLM call fails. Even on the regex
// fallback path, LLM is still used to extract base names for highlighting.
func parseIngredientList(rawList []string, llm langllms.Model) []utils.Ingredient {
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
	if llm != nil {
		slog.Info("Attempting LLM ingredient parsing", "count", len(cleaned))
		parsed, err := utils.ParseIngredientsWithLLM(context.Background(), cleaned, llm)
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
	if llm != nil && len(ingredients) > 0 {
		slog.Info("Attempting LLM base name extraction", "count", len(ingredients))
		if err := utils.ExtractBaseNamesWithLLM(context.Background(), ingredients, llm); err != nil {
			slog.Error("LLM base name extraction failed, using full names for highlighting", "error", err)
		}
	}

	return ingredients
}
