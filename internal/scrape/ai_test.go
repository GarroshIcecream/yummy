package scrape

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/tmc/langchaingo/llms"
)

const samplePage = `<!doctype html><html><head><title>Grandma's Pie</title>
<style>body{color:red}</style>
<script>var tracking = "do not include";</script>
<script type="application/ld+json">{"@type":"Recipe","name":"Grandma's Pie"}</script>
</head><body>
<nav>Home | Recipes | About</nav>
<main><h1>Grandma's Pie</h1>
<ul><li>2 cups flour</li><li>1 tsp salt</li></ul>
<p>Mix   and
bake.</p></main>
<footer>© 2026 Cookie policy</footer>
</body></html>`

func TestFetchPageTextKeepsContentAndJSONLD(t *testing.T) {
	var gotUA string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
		_, _ = w.Write([]byte(samplePage))
	}))
	defer srv.Close()

	text, err := FetchPageText(context.Background(), srv.Client(), srv.URL)
	if err != nil {
		t.Fatalf("FetchPageText: %v", err)
	}
	for _, want := range []string{`"@type":"Recipe"`, "Grandma's Pie", "2 cups flour", "1 tsp salt", "Mix and bake."} {
		if !strings.Contains(text, want) {
			t.Errorf("text is missing %q:\n%s", want, text)
		}
	}
	for _, unwanted := range []string{"tracking", "color:red", "Cookie policy", "About"} {
		if strings.Contains(text, unwanted) {
			t.Errorf("text should not contain %q:\n%s", unwanted, text)
		}
	}
	if gotUA == "" || strings.Contains(gotUA, "Go-http-client") {
		t.Errorf("User-Agent = %q, want a browser-like agent", gotUA)
	}
}

func TestFetchPageTextHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusNotFound)
	}))
	defer srv.Close()
	if _, err := FetchPageText(context.Background(), srv.Client(), srv.URL); err == nil || !strings.Contains(err.Error(), "404") {
		t.Fatalf("err = %v, want a 404 error", err)
	}
}

func TestFetchPageTextTruncatesLongPages(t *testing.T) {
	long := "<html><body><p>" + strings.Repeat("word ", maxPageChars) + "</p></body></html>"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(long))
	}))
	defer srv.Close()
	text, err := FetchPageText(context.Background(), srv.Client(), srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	if len(text) > maxPageChars {
		t.Errorf("len(text) = %d, want <= %d", len(text), maxPageChars)
	}
}

type fakeLLM struct {
	reply     string
	err       error
	prompt    string
	jsonMode  bool
	maxTokens int
}

func (f *fakeLLM) GenerateContent(_ context.Context, msgs []llms.MessageContent, opts ...llms.CallOption) (*llms.ContentResponse, error) {
	var o llms.CallOptions
	for _, opt := range opts {
		opt(&o)
	}
	f.jsonMode = o.JSONMode
	f.maxTokens = o.MaxTokens
	for _, m := range msgs {
		for _, p := range m.Parts {
			if tp, ok := p.(llms.TextContent); ok {
				f.prompt += tp.Text + "\n"
			}
		}
	}
	if f.err != nil {
		return nil, f.err
	}
	return &llms.ContentResponse{Choices: []*llms.ContentChoice{{Content: f.reply}}}, nil
}

func (f *fakeLLM) Call(ctx context.Context, prompt string, opts ...llms.CallOption) (string, error) {
	return llms.GenerateFromSinglePrompt(ctx, f, prompt, opts...)
}

func TestExtractRecipeWithLLM(t *testing.T) {
	f := &fakeLLM{reply: "```json\n" + `{
		"found": true,
		"title": "Grandma's Pie",
		"description": "Flaky.",
		"author": "Grandma",
		"yields": "8 servings",
		"prep_time": 20, "cook_time": 45, "total_time": 65,
		"ingredient_groups": [
			{"purpose": "Crust", "ingredients": ["2 cups flour", "1 tsp salt"]},
			{"purpose": "Filling", "ingredients": ["6 apples"]}
		],
		"instructions": ["Mix.", "Bake."],
		"categories": ["Dessert"],
		"cuisine": ["American"]
	}` + "\n```"}
	s, err := ExtractRecipeWithLLM(context.Background(), f, "https://pie.example/r/1", "PAGE TEXT")
	if err != nil {
		t.Fatalf("ExtractRecipeWithLLM: %v", err)
	}
	if !f.jsonMode {
		t.Error("request did not use JSON mode")
	}
	if !strings.Contains(f.prompt, "PAGE TEXT") || !strings.Contains(f.prompt, "https://pie.example/r/1") {
		t.Errorf("prompt is missing the page or URL:\n%s", f.prompt)
	}
	if name, _ := s.Name(); name != "Grandma's Pie" {
		t.Errorf("Name = %q", name)
	}
	if ings, ok := s.Ingredients(); !ok || len(ings) != 3 || ings[2] != "6 apples" {
		t.Errorf("Ingredients = %v, %v", ings, ok)
	}
	if groups, ok := s.IngredientGroups(); !ok || len(groups) != 2 || *groups[0].Purpose != "Crust" {
		t.Errorf("IngredientGroups = %#v", groups)
	}
	if steps, ok := s.Instructions(); !ok || len(steps) != 2 {
		t.Errorf("Instructions = %v", steps)
	}
	if d, ok := s.CookTime(); !ok || d != 45*time.Minute {
		t.Errorf("CookTime = %v", d)
	}
	if y, _ := s.Yields(); y != "8 servings" {
		t.Errorf("Yields = %q", y)
	}
	if cats, _ := s.Categories(); len(cats) != 1 || cats[0] != "Dessert" {
		t.Errorf("Categories = %v", cats)
	}
	if host, _ := s.Host(); host != "pie.example" {
		t.Errorf("Host = %q", host)
	}
}

func TestExtractRecipeWithLLMNoRecipe(t *testing.T) {
	for _, reply := range []string{`{"found": false}`, `{"found": true, "title": "Pie", "ingredient_groups": []}`} {
		_, err := ExtractRecipeWithLLM(context.Background(), &fakeLLM{reply: reply}, "https://x.example", "text")
		if !errors.Is(err, ErrNoRecipeOnPage) {
			t.Errorf("reply %s: err = %v, want ErrNoRecipeOnPage", reply, err)
		}
	}
}

func TestExtractRecipeWithLLMPassesModelError(t *testing.T) {
	boom := errors.New("quota exceeded")
	if _, err := ExtractRecipeWithLLM(context.Background(), &fakeLLM{err: boom}, "https://x.example", "text"); !errors.Is(err, boom) {
		t.Fatalf("err = %v, want it to wrap %v", err, boom)
	}
}

func TestFetchPageTextReportsBlockedSite(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("cf-mitigated", "challenge")
		http.Error(w, "Just a moment...", http.StatusForbidden)
	}))
	defer srv.Close()
	_, err := FetchPageText(context.Background(), srv.Client(), srv.URL)
	if !errors.Is(err, ErrSiteBlocked) {
		t.Fatalf("err = %v, want ErrSiteBlocked", err)
	}
}

func TestExtractRecipeFromURLWithLLM(t *testing.T) {
	f := &fakeLLM{reply: "Here it is:\n```json\n" + `{"found": true, "title": "Ramen", "ingredient_groups": [{"purpose": "", "ingredients": ["2 packs noodles"]}], "instructions": ["Boil."]}` + "\n```"}
	marker := llms.WithMaxTokens(7) // stands in for the provider's web tool option
	s, err := ExtractRecipeFromURLWithLLM(context.Background(), f, "https://blocked.example/ramen", marker)
	if err != nil {
		t.Fatalf("ExtractRecipeFromURLWithLLM: %v", err)
	}
	if f.jsonMode {
		t.Error("JSON mode should be off: it can't be combined with the web tool")
	}
	if f.maxTokens != 7 {
		t.Error("caller's options (the web tool) were not passed to the model")
	}
	if !strings.Contains(f.prompt, "https://blocked.example/ramen") {
		t.Errorf("prompt is missing the URL:\n%s", f.prompt)
	}
	if name, _ := s.Name(); name != "Ramen" {
		t.Errorf("Name = %q", name)
	}
}

func TestExtractRecipeFromURLWithLLMCouldNotOpen(t *testing.T) {
	_, err := ExtractRecipeFromURLWithLLM(context.Background(), &fakeLLM{reply: `{"found": false}`}, "https://blocked.example/ramen")
	if !errors.Is(err, ErrSiteBlocked) {
		t.Fatalf("err = %v, want ErrSiteBlocked (the model couldn't open it either)", err)
	}
}
