package dialog

import (
	"errors"
	"strings"
	"testing"

	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	themes "github.com/GarroshIcecream/yummy/internal/themes"
	"github.com/charmbracelet/x/ansi"
)

func newTestAddRecipeDialog(aiFallback bool) *AddRecipeFromURLDialogCmp {
	theme := themes.NewDefaultTheme()
	ti := textinput.New()
	ti.Focus()
	ti.SetValue("https://blog.example/pie")
	d := &AddRecipeFromURLDialogCmp{
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
