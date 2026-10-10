package utils

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
)

func TestFillBackgroundUniformWidthAndColor(t *testing.T) {
	bg := lipgloss.Color("#1D2330")

	// A block mixing a styled span (emits a reset), trailing plain text, a
	// short line, and a longer line — the exact shape that renders ragged.
	span := lipgloss.NewStyle().Background(lipgloss.Color("#FF0000")).Render("HI")
	block := lipgloss.JoinVertical(lipgloss.Left,
		span+"  tail",
		"short",
		"a noticeably longer line",
	)

	const width = 30
	out := FillBackground(block, bg, width)

	bgSeq := "\x1b[48;2;29;35;48m" // #1D2330
	for i, line := range strings.Split(out, "\n") {
		if w := lipgloss.Width(line); w != width {
			t.Errorf("line %d: visible width = %d, want %d (%q)", i, w, width, line)
		}
		if !strings.HasPrefix(line, bgSeq) {
			t.Errorf("line %d: does not begin with background sequence (%q)", i, line)
		}
		// After every reset the background must be re-established, so no reset
		// may be immediately followed by a newline-bound bare region. The
		// simplest invariant: the line must not contain a reset that is the
		// final SGR before padding without bg following it.
		if strings.Contains(line, "\x1b[m") && !strings.Contains(line, "\x1b[m\x1b[48;2;29;35;48m") &&
			!strings.HasSuffix(line, "\x1b[m") {
			t.Errorf("line %d: reset not followed by background (%q)", i, line)
		}
	}
}

func TestFillBackgroundNilIsNoop(t *testing.T) {
	if got := FillBackground("abc", nil, 10); got != "abc" {
		t.Errorf("nil bg should be a no-op, got %q", got)
	}
}

func TestFillBackgroundHandlesDefaultBgAndCompoundResets(t *testing.T) {
	bg := lipgloss.Color("#1D2330")
	bgSeq := "\x1b[48;2;29;35;48m"

	cases := map[string]bool{
		"\x1b[49mx":          true,  // default background
		"\x1b[39;49mx":       true,  // compound fg+bg default
		"\x1b[0;1mx":         true,  // reset inside compound
		"\x1b[1mx":           false, // bold only — bg untouched
		"\x1b[38;5;0mx":      false, // 0 is a color index, not a reset
		"\x1b[38;2;0;49;0mx": false, // 0 / 49 are channel values
		"\x1b[0;48;5;3mx":    false, // reset then explicit bg
	}
	for in, wantFill := range cases {
		out := FillBackground(in, bg, 0)
		// Strip the leading bracket sequence the function always adds.
		body := strings.TrimPrefix(out, bgSeq)
		got := strings.Contains(body, bgSeq)
		if got != wantFill {
			t.Errorf("%q: re-established bg = %v, want %v (out %q)", in, got, wantFill, out)
		}
	}
}
