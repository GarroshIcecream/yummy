package scrape

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"unicode/utf8"

	"github.com/tmc/langchaingo/llms"
	"golang.org/x/net/html"
)

// maxPageChars caps the page text sent to the model, so a huge page doesn't
// blow the context window (or the user's quota).
const maxPageChars = 60000

// maxPageBytes caps how much HTML is downloaded.
const maxPageBytes = 5 << 20

// browserUserAgent is sent because many recipe sites refuse Go's default.
const browserUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0 Safari/537.36"

// ErrSiteBlocked means the site refused the download, typically a bot
// check (e.g. a Cloudflare challenge) that only a real browser passes.
var ErrSiteBlocked = errors.New("site blocks automated downloads")

// ErrNoRecipeOnPage means the model found no recipe in the page.
var ErrNoRecipeOnPage = errors.New("no recipe found on this page")

// skippedElements hold no recipe content (code, styling, site chrome).
var skippedElements = map[string]bool{
	"script": true, "style": true, "noscript": true, "template": true,
	"svg": true, "iframe": true, "nav": true, "footer": true, "form": true,
	"button": true, "select": true,
}

// blockElements end a line in the extracted text.
var blockElements = map[string]bool{
	"p": true, "div": true, "li": true, "br": true, "tr": true, "section": true,
	"article": true, "main": true, "header": true, "h1": true, "h2": true,
	"h3": true, "h4": true, "h5": true, "h6": true, "ul": true, "ol": true,
	"table": true, "dt": true, "dd": true, "figcaption": true, "blockquote": true,
}

// FetchPageText downloads pageURL and returns its readable text, with any
// schema.org JSON-LD first (it often holds the whole recipe). The result is
// at most maxPageChars long.
func FetchPageText(ctx context.Context, client *http.Client, pageURL string) (string, error) {
	if client == nil {
		client = http.DefaultClient
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pageURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", browserUserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml;q=0.9,*/*;q=0.8")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusUnauthorized ||
		resp.Header.Get("cf-mitigated") != "" {
		return "", fmt.Errorf("%w (%s)", ErrSiteBlocked, resp.Status)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("page returned %s", resp.Status)
	}
	doc, err := html.Parse(io.LimitReader(resp.Body, maxPageBytes))
	if err != nil {
		return "", fmt.Errorf("parse page: %w", err)
	}

	var jsonLD []string
	var text strings.Builder
	var walk func(n *html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode {
			if n.Data == "script" && attr(n, "type") == "application/ld+json" {
				if n.FirstChild != nil {
					jsonLD = append(jsonLD, strings.TrimSpace(n.FirstChild.Data))
				}
				return
			}
			if skippedElements[n.Data] {
				return
			}
		}
		if n.Type == html.TextNode {
			if t := strings.Join(strings.Fields(n.Data), " "); t != "" {
				if text.Len() > 0 && !strings.HasSuffix(text.String(), "\n") {
					text.WriteByte(' ')
				}
				text.WriteString(t)
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
		if n.Type == html.ElementNode && blockElements[n.Data] && text.Len() > 0 &&
			!strings.HasSuffix(text.String(), "\n") {
			text.WriteByte('\n')
		}
	}
	walk(doc)

	var out strings.Builder
	if len(jsonLD) > 0 {
		out.WriteString("Structured data (JSON-LD):\n")
		out.WriteString(strings.Join(jsonLD, "\n"))
		out.WriteString("\n\nPage text:\n")
	}
	out.WriteString(strings.TrimSpace(text.String()))
	return truncateUTF8(out.String(), maxPageChars), nil
}

func attr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return strings.ToLower(strings.TrimSpace(a.Val))
		}
	}
	return ""
}

func truncateUTF8(s string, n int) string {
	if len(s) <= n {
		return s
	}
	s = s[:n]
	// Drop a rune cut in half at the end.
	for len(s) > 0 && !utf8.ValidString(s) {
		s = s[:len(s)-1]
	}
	return s
}

const extractRecipePrompt = `You extract a cooking recipe from a web page. Reply with one JSON object and nothing else, using exactly these keys:
{
  "found": true if the page contains a recipe, otherwise false,
  "title": string,
  "description": string (one or two sentences, "" if none),
  "author": string ("" if unknown),
  "yields": string, e.g. "4 servings" ("" if unknown),
  "prep_time": minutes as an integer (0 if unknown),
  "cook_time": minutes as an integer (0 if unknown),
  "total_time": minutes as an integer (0 if unknown),
  "ingredient_groups": [{"purpose": string ("" when the recipe has a single group), "ingredients": [string, ...]}],
  "instructions": [string, ...] (one step per item),
  "categories": [string, ...] (e.g. "Dessert"; [] if none),
  "cuisine": [string, ...] ([] if none)
}
Copy ingredient lines and steps as written on the page, in the page's language, including amounts and units. Do not invent anything that is not on the page. Ignore comments, ads and related recipes.`

type aiRecipe struct {
	Found            bool   `json:"found"`
	Title            string `json:"title"`
	Description      string `json:"description"`
	Author           string `json:"author"`
	Yields           string `json:"yields"`
	PrepTime         int    `json:"prep_time"`
	CookTime         int    `json:"cook_time"`
	TotalTime        int    `json:"total_time"`
	IngredientGroups []struct {
		Purpose     string   `json:"purpose"`
		Ingredients []string `json:"ingredients"`
	} `json:"ingredient_groups"`
	Instructions []string `json:"instructions"`
	Categories   []string `json:"categories"`
	Cuisine      []string `json:"cuisine"`
}

// ExtractRecipeWithLLM asks llm to read the recipe out of pageText (as
// returned by FetchPageText) and returns it as a Scraper, like a
// recipe-scrapers result. It returns ErrNoRecipeOnPage when the model finds
// no recipe.
func ExtractRecipeWithLLM(ctx context.Context, llm llms.Model, pageURL, pageText string) (Scraper, error) {
	msgs := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, extractRecipePrompt),
		llms.TextParts(llms.ChatMessageTypeHuman, "URL: "+pageURL+"\n\n"+pageText),
	}
	resp, err := llm.GenerateContent(ctx, msgs, llms.WithJSONMode())
	if err != nil {
		return nil, fmt.Errorf("read recipe with AI: %w", err)
	}
	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("read recipe with AI: empty response")
	}
	return parseAIRecipe(resp.Choices[0].Content, pageURL)
}

// parseAIRecipe turns the model's JSON reply into a Scraper. The JSON may be
// wrapped in a code fence or surrounded by prose.
func parseAIRecipe(reply, pageURL string) (Scraper, error) {
	raw := stripCodeFence(reply)
	if i, j := strings.IndexByte(raw, '{'), strings.LastIndexByte(raw, '}'); i >= 0 && j > i {
		raw = raw[i : j+1]
	}
	var r aiRecipe
	if err := json.Unmarshal([]byte(raw), &r); err != nil {
		return nil, fmt.Errorf("read recipe with AI: invalid JSON: %w", err)
	}

	j := RecipeScrapersJSON{
		Title:            strings.TrimSpace(r.Title),
		Description:      r.Description,
		Author:           r.Author,
		Yields:           r.Yields,
		PrepTime:         positive(r.PrepTime),
		CookTime:         positive(r.CookTime),
		TotalTime:        positive(r.TotalTime),
		InstructionsList: nonEmpty(r.Instructions),
		Category:         FlexibleString(strings.Join(r.Categories, ", ")),
		Cuisine:          FlexibleString(strings.Join(r.Cuisine, ", ")),
		CanonicalURL:     pageURL,
	}
	if u, err := url.Parse(pageURL); err == nil {
		j.Host = u.Hostname()
	}
	for _, g := range r.IngredientGroups {
		ings := nonEmpty(g.Ingredients)
		if len(ings) == 0 {
			continue
		}
		j.Ingredients = append(j.Ingredients, ings...)
		purpose := strings.TrimSpace(g.Purpose)
		j.IngredientGroups = append(j.IngredientGroups, IngredientGroup{Ingredients: ings, Purpose: &purpose})
	}
	if !r.Found || j.Title == "" || len(j.Ingredients) == 0 {
		return nil, ErrNoRecipeOnPage
	}
	if len(j.IngredientGroups) == 1 && *j.IngredientGroups[0].Purpose == "" {
		j.IngredientGroups = nil // a single unnamed group is just the ingredient list
	}
	return &adapter{j: j}, nil
}

const extractFromURLPrompt = `Open the web page at the URL below with your web tool and extract the cooking recipe on that exact page. Use only that page: if you can't open it, or it has no recipe, reply {"found": false}. Do not substitute a similar recipe from another site.

`

// ExtractRecipeFromURLWithLLM has the model open pageURL itself, through a
// hosted web tool passed in opts (e.g. chatgpt.WithWebSearch), for pages
// that refuse direct downloads. It returns an error wrapping ErrSiteBlocked
// when the model can't open the page or finds no recipe there.
func ExtractRecipeFromURLWithLLM(ctx context.Context, llm llms.Model, pageURL string, opts ...llms.CallOption) (Scraper, error) {
	msgs := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, extractRecipePrompt),
		llms.TextParts(llms.ChatMessageTypeHuman, extractFromURLPrompt+"URL: "+pageURL),
	}
	// No JSON mode: OpenAI doesn't allow it together with the web tool, so
	// the prompt asks for JSON and parseAIRecipe digs it out of the reply.
	resp, err := llm.GenerateContent(ctx, msgs, opts...)
	if err != nil {
		return nil, fmt.Errorf("read recipe with AI: %w", err)
	}
	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("read recipe with AI: empty response")
	}
	s, err := parseAIRecipe(resp.Choices[0].Content, pageURL)
	if errors.Is(err, ErrNoRecipeOnPage) {
		return nil, fmt.Errorf("%w: the AI couldn't open it either", ErrSiteBlocked)
	}
	return s, err
}

func stripCodeFence(s string) string {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "```") {
		return s
	}
	s = strings.TrimPrefix(s, "```")
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		s = s[i+1:]
	}
	return strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(s), "```"))
}

func positive(n int) *int {
	if n <= 0 {
		return nil
	}
	return &n
}

func nonEmpty(in []string) []string {
	var out []string
	for _, s := range in {
		if t := strings.TrimSpace(s); t != "" {
			out = append(out, t)
		}
	}
	return out
}
