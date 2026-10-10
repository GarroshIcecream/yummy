package markdown

import (
	"strings"
	"testing"

	"charm.land/glamour/v2"
	"charm.land/lipgloss/v2"
	themes "github.com/GarroshIcecream/yummy/internal/themes"
)

// TestThemedRendererUsesThemePalette verifies the shared renderer applies the
// app's themed StyleConfig (headings in the accent blue #4A9EFF) rather than
// glamour's environment default, so every markdown surface matches.
func TestThemedRendererUsesThemePalette(t *testing.T) {
	r, err := NewThemedRenderer(nil, 60)
	if err != nil {
		t.Fatalf("NewThemedRenderer: %v", err)
	}
	out, err := r.Render("# Heading\n\nBody text.")
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if !strings.Contains(out, "38;2;74;158;255") { // #4A9EFF
		t.Errorf("themed heading color (#4A9EFF) not applied; output=%q", out)
	}
}

// TestThemedRendererAcceptsExtraOptions ensures caller extras (e.g. emoji) are
// accepted without error — the detail/cooking views pass glamour.WithEmoji().
func TestThemedRendererAcceptsExtraOptions(t *testing.T) {
	if _, err := NewThemedRenderer(nil, 40, glamour.WithEmoji()); err != nil {
		t.Fatalf("NewThemedRenderer with extra option: %v", err)
	}
}

// TestThemedRendererFollowsTheme verifies headings take the active theme's
// primary color, so switching themes restyles markdown too.
func TestThemedRendererFollowsTheme(t *testing.T) {
	p := themes.DefaultPalette()
	p.Primary = lipgloss.Color("#123456")
	theme := themes.NewThemeFromPalette("t", "", p)
	r, err := NewThemedRenderer(&theme, 60)
	if err != nil {
		t.Fatalf("NewThemedRenderer: %v", err)
	}
	out, err := r.Render("# Heading\n\nBody text.")
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if !strings.Contains(out, "38;2;18;52;86") {
		t.Errorf("heading should use the theme primary #123456; output=%q", out)
	}
	if strings.Contains(out, "# ") {
		t.Errorf("headings should not render literal '#' markers; output=%q", out)
	}
}

// TestRendererEmitsHyperlinks guards that URLs are rendered as OSC 8
// hyperlinks (clickable in supporting terminals once mouse capture is off).
func TestRendererEmitsHyperlinks(t *testing.T) {
	r, err := NewThemedRenderer(nil, 80)
	if err != nil {
		t.Fatalf("NewThemedRenderer: %v", err)
	}
	out, err := r.Render("Source: https://example.com/recipe\n")
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if !strings.Contains(out, "\x1b]8;") || !strings.Contains(out, ";https://example.com/recipe\a") {
		t.Errorf("expected an OSC 8 hyperlink; output=%q", out)
	}
}
