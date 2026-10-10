package chat

import (
	"strings"
	"testing"

	themes "github.com/GarroshIcecream/yummy/internal/themes"
)

// mentionFgSGR is the SGR foreground for the default theme's ChatMention color
// (the palette primary). Used to assert the accent actually wraps a mention.
const mentionFgSGR = "38;2;74;158;255" // palette primary #4A9EFF

// stripSGR removes ANSI SGR sequences so assertions can inspect visible text.
func stripSGR(s string) string { return sgrRe.ReplaceAllString(s, "") }

func TestHighlightMentions_NoMention(t *testing.T) {
	th := themes.NewDefaultTheme()
	in := "just some plain text"
	if got := HighlightMentions(in, &th); got != in {
		t.Fatalf("expected unchanged content, got %q", got)
	}
}

// TestHighlightMentions_GlamourInterleavedANSI is the regression test for the
// bug where glamour injects reset codes inside "@[Name]" (e.g.
// "@[\x1b[m...Baklava]"), which previously neutralized the highlight so
// mentions rendered in plain text.
func TestHighlightMentions_GlamourInterleavedANSI(t *testing.T) {
	r, err := newThemedMarkdownRenderer(nil, 80)
	if err != nil {
		t.Fatal(err)
	}
	rendered, err := r.Render("@[Baklava] what is this recipe man")
	if err != nil {
		t.Fatal(err)
	}

	th := themes.NewDefaultTheme()
	out := HighlightMentions(rendered, &th)

	// The mention is displayed as a clean "@Baklava" token (the internal
	// "[...]" markup is stripped), and the accent color must wrap it
	// contiguously with no reset between the opening accent SGR and the token.
	idx := strings.Index(out, mentionFgSGR)
	if idx < 0 {
		t.Fatalf("mention accent color not applied; output=%q", out)
	}
	afterAccent := out[idx:]
	if !strings.Contains(stripSGR(afterAccent[:len("@Baklava")+40]), "@Baklava") {
		t.Fatalf("accent color does not wrap the mention token; output=%q", out)
	}
	// The bracket markup must not survive into the displayed text.
	if strings.Contains(stripSGR(out), "@[Baklava]") {
		t.Fatalf("raw bracket markup leaked into display; output=%q", out)
	}

	// The accent run must not be broken by a reset inside the mention text.
	accentToReset := afterAccent[:strings.Index(afterAccent, "\x1b[m")]
	if !strings.Contains(stripSGR(accentToReset), "@Baklava") {
		t.Fatalf("reset code splits the mention; output=%q", out)
	}

	// Trailing text must remain present (and re-colored, not stripped).
	if !strings.Contains(stripSGR(out), "what is this recipe man") {
		t.Fatalf("trailing text lost; output=%q", out)
	}
}
