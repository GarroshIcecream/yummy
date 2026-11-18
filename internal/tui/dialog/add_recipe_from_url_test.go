package dialog

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/GarroshIcecream/yummy/internal/config"
	"github.com/GarroshIcecream/yummy/internal/scrape"
	themes "github.com/GarroshIcecream/yummy/internal/themes"
	"github.com/charmbracelet/x/ansi"
)

func newTestAddRecipeDialog(aiFallback bool) *AddRecipeFromURLDialogCmp {
	theme := themes.NewDefaultTheme()
	ti := textinput.New()
	ti.Focus()
	ti.SetValue("https://blog.example/pie")
	km := config.NewDefaultKeyBindings()
	km.CommandPalette = []string{"ctrl+k"}
	d := &AddRecipeFromURLDialogCmp{
		keyMap:     km.ToKeyMap(),
		urlInput:   ti,
		spinner:    spinner.New(),
		theme:      &theme,
		width:      70,
		aiFallback: aiFallback,
		fetchedURL: "https://blog.example/pie",
		loading:    true,
	}
	return d
}

func viewText(d *AddRecipeFromURLDialogCmp) string { return ansi.Strip(d.View().Content) }

func TestAddRecipeOffersAIWhenScrapingUnsupported(t *testing.T) {
	d := newTestAddRecipeDialog(true)
	d.Update(scrapeFailedMsg{err: errors.New("Website (blog.example) isn't currently supported by recipe-scrapers!")})
	if !d.aiOffer {
		t.Fatal("no AI offer after an unsupported-site error")
	}
	if v := viewText(d); !strings.Contains(v, "Can't scrape this website") || !strings.Contains(v, "ChatGPT") {
		t.Errorf("view should explain the failure and offer ChatGPT:\n%s", v)
	}
	_, cmd := d.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if !d.loading || cmd == nil {
		t.Fatalf("enter did not start the AI import (loading=%v, cmd=%v)", d.loading, cmd != nil)
	}
	if d.aiOffer {
		t.Error("offer should be consumed once the AI import starts")
	}
	if !strings.Contains(viewText(d), "ChatGPT") {
		t.Errorf("spinner should say ChatGPT is reading the page:\n%s", viewText(d))
	}
}

func TestAddRecipeNoAIOfferWithoutChatGPT(t *testing.T) {
	d := newTestAddRecipeDialog(false)
	d.Update(scrapeFailedMsg{err: errors.New("Website (blog.example) isn't currently supported by recipe-scrapers!")})
	if d.aiOffer {
		t.Fatal("AI offered while the provider isn't ChatGPT")
	}
	if strings.Contains(viewText(d), "ChatGPT") {
		t.Errorf("view mentions ChatGPT:\n%s", viewText(d))
	}
}

func TestAddRecipeNoAIOfferWhenSiteUnreachable(t *testing.T) {
	for _, msg := range []string{"dial tcp: lookup blog.example: no such host", "HTTP Error 404: Not Found"} {
		d := newTestAddRecipeDialog(true)
		d.Update(scrapeFailedMsg{err: errors.New(msg)})
		if d.aiOffer {
			t.Errorf("%q: AI offered for a page that can't be fetched", msg)
		}
	}
}

func TestAddRecipeOffersAIWithoutPython(t *testing.T) {
	d := newTestAddRecipeDialog(true)
	d.Update(scrapeFailedMsg{err: errors.New("no Python found in PATH (try setting python_path in config)")})
	if !d.aiOffer {
		t.Fatal("AI import doesn't need Python, so it should be offered")
	}
}

func TestAddRecipeEditingURLWithdrawsAIOffer(t *testing.T) {
	d := newTestAddRecipeDialog(true)
	d.Update(scrapeFailedMsg{err: errors.New("No recipe found")})
	d.Update(tea.KeyPressMsg{Code: 'x', Text: "x"})
	if d.aiOffer {
		t.Fatal("offer should be withdrawn once the URL changes")
	}
}

func TestFriendlyScrapeErrorHidesRawOutput(t *testing.T) {
	cases := map[string]string{
		"exit status 1 (stderr: HTTP Error 403: Forbidden\n)":         "This site blocked the import",
		"exit status 1 (stderr: HTTP Error 429: Too Many Requests\n)": "This site is rate limiting imports, try again later",
		"exit status 1 (stderr: HTTP Error 503: Service Unavailable)": "This site is having trouble, try again later",
		"exit status 1 (stderr: Traceback ... KeyError: 'whatever')":  "Couldn't read a recipe from this page",
	}
	for raw, want := range cases {
		if got := friendlyScrapeError(errors.New(raw)); got != want {
			t.Errorf("friendlyScrapeError(%q) = %q, want %q", raw, got, want)
		}
	}
}

func TestAddRecipeAIOfferIsTheHeadline(t *testing.T) {
	d := newTestAddRecipeDialog(true)
	d.Update(scrapeFailedMsg{err: errors.New("exit status 1 (stderr: HTTP Error 403: Forbidden\n)")})
	if !d.aiOffer {
		t.Fatal("a site blocking recipe-scrapers should still get the AI offer")
	}
	v := viewText(d)
	if strings.Contains(v, "exit status") || strings.Contains(v, "stderr") {
		t.Errorf("raw scraper output leaked into the dialog:\n%s", v)
	}
	if !strings.Contains(v, "This site blocked the import") {
		t.Errorf("view should give the short reason:\n%s", v)
	}
}

func TestAddRecipeAIImportBlockedSaysWhy(t *testing.T) {
	d := newTestAddRecipeDialog(true)
	d.Update(scrapeAndSaveResultMsg{err: fmt.Errorf("%w (403 Forbidden)", scrape.ErrSiteBlocked)})
	v := viewText(d)
	if !strings.Contains(v, "blocks automated downloads") || !strings.Contains(v, "ChatGPT can't read it") {
		t.Errorf("view should explain the site blocks downloads, not that ChatGPT failed:\n%s", v)
	}
}

// lineWith returns the index of the first view line containing s, or -1.
func lineWith(view, s string) int {
	for i, l := range strings.Split(view, "\n") {
		if strings.Contains(l, s) {
			return i
		}
	}
	return -1
}

func TestAddRecipeHelpWrapsBetweenItems(t *testing.T) {
	d := newTestAddRecipeDialog(true)
	d.SetSize(56, 30) // narrow terminal
	d.Update(scrapeFailedMsg{err: errors.New("exit status 1 (stderr: HTTP Error 403: Forbidden\n)")})
	v := viewText(d)
	for _, item := range []string{"enter use ChatGPT", "tab write manually", "esc cancel"} {
		if lineWith(v, item) < 0 {
			t.Errorf("help item %q is split across lines:\n%s", item, v)
		}
	}
}

func TestAddRecipeReasonThenOffer(t *testing.T) {
	d := newTestAddRecipeDialog(true)
	d.Update(scrapeFailedMsg{err: errors.New("exit status 1 (stderr: HTTP Error 403: Forbidden\n)")})
	v := viewText(d)
	reason, offer := lineWith(v, "This site blocked the import"), lineWith(v, "ChatGPT read the page")
	if reason < 0 || offer < 0 || reason > offer {
		t.Errorf("want the reason, then the offer (reason line %d, offer line %d):\n%s", reason, offer, v)
	}
}

func TestAddRecipeAIHintUsesPaletteKey(t *testing.T) {
	d := newTestAddRecipeDialog(false)
	d.loading = false
	if v := viewText(d); !strings.Contains(v, "ctrl+k") || strings.Contains(v, "ctrl+p") {
		t.Errorf("AI parsing hint should name the configured palette key (ctrl+k):\n%s", v)
	}
}

func TestAddRecipeLoadingShowsOnlyCancel(t *testing.T) {
	d := newTestAddRecipeDialog(true) // starts mid-import
	v := viewText(d)
	if !strings.Contains(v, "esc cancel") || strings.Contains(v, "write manually") {
		t.Errorf("mid-import help should only offer esc:\n%s", v)
	}
}
