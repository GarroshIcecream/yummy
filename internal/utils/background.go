package utils

import (
	"fmt"
	"image/color"
	"regexp"
	"strconv"
	"strings"

	"charm.land/lipgloss/v2"
)

// FillBackground makes a rendered (possibly multi-line, ANSI-styled) block sit
// on a uniform background color.
//
// It exists because lipgloss does not re-establish a container's background
// after the SGR resets that every styled span emits: on any composed line,
// the cells following the first styled span — and the padding lipgloss.JoinVertical
// inserts between lines of differing width — fall through to the terminal's own
// background, producing ragged, patchy bands.
//
// For every line FillBackground:
//   - re-establishes bg immediately after every SGR that clears it — a full
//     reset (\x1b[0m / \x1b[m, also inside compound sequences) or a
//     default-background (\x1b[49m, which glamour emits after every span) — so
//     cells after a styled span keep the background;
//   - right-pads the line to width visible cells (when width > 0) so short
//     lines extend the background to the full region width;
//   - brackets the line with a leading background sequence and a trailing reset,
//     so every cell of every line is explicitly colored regardless of how the
//     terminal treats line boundaries.
//
// Spans that set their own background (e.g. a themed text input) keep it; only
// the otherwise-unstyled cells inherit bg. A nil bg returns s unchanged.
func FillBackground(s string, bg color.Color, width int) string {
	return FillColors(s, nil, bg, width)
}

// FillColors is FillBackground plus a default foreground: after every SGR that
// resets the foreground (a full reset or \x1b[39m) fg is re-established, so
// unstyled text renders in the theme's text color instead of the terminal's.
// A nil fg behaves exactly like FillBackground; a nil bg returns s unchanged.
func FillColors(s string, fg, bg color.Color, width int) string {
	if bg == nil {
		return s
	}

	bgSeq := sgrColor(48, bg)
	fgSeq := ""
	if fg != nil {
		fgSeq = sgrColor(38, fg)
	}

	lines := strings.Split(s, "\n")
	for i, line := range lines {
		line = sgrRe.ReplaceAllStringFunc(line, func(m string) string {
			clearsFg, clearsBg := sgrClears(m[2 : len(m)-1])
			if clearsBg {
				m += bgSeq
			}
			if clearsFg && fgSeq != "" {
				m += fgSeq
			}
			return m
		})
		if width > 0 {
			if pad := width - lipgloss.Width(line); pad > 0 {
				line += strings.Repeat(" ", pad)
			}
		}
		lines[i] = bgSeq + fgSeq + line + "\x1b[0m"
	}
	return strings.Join(lines, "\n")
}

func sgrColor(code int, c color.Color) string {
	r, g, b, _ := c.RGBA()
	// RGBA returns 16-bit channels; scale to 8-bit for the SGR sequence.
	return fmt.Sprintf("\x1b[%d;2;%d;%d;%dm", code, r>>8, g>>8, b>>8)
}

// sgrRe matches a single SGR (Select Graphic Rendition) escape sequence.
var sgrRe = regexp.MustCompile("\x1b\\[[0-9;:]*m")

// sgrClears reports whether an SGR parameter list leaves the foreground and/or
// background at the terminal default: a full reset (empty or 0), an explicit
// default-foreground (39) or default-background (49). A later explicit color in
// the same sequence wins. Extended color arguments (38/48/58 followed by 5;n or
// 2;r;g;b) are skipped so a literal 0 or 49 inside a color value does not count.
func sgrClears(params string) (fg, bg bool) {
	if params == "" {
		return true, true
	}
	parts := strings.Split(params, ";")
	for i := 0; i < len(parts); i++ {
		p := parts[i]
		if strings.Contains(p, ":") {
			// Colon sub-parameters (e.g. 48:2::r:g:b) are self-contained.
			if strings.HasPrefix(p, "38:") {
				fg = false
			} else if strings.HasPrefix(p, "48:") {
				bg = false
			}
			continue
		}
		switch p {
		case "", "0":
			fg, bg = true, true
		case "39":
			fg = true
		case "49":
			bg = true
		case "38", "48", "58":
			if i+1 < len(parts) {
				switch parts[i+1] {
				case "5":
					i += 2
				case "2":
					i += 4
				}
			}
			switch p {
			case "38":
				fg = false
			case "48":
				bg = false
			}
		default:
			// Basic / bright palette colors (30–37, 90–97 fg; 40–47, 100–107 bg).
			if n, err := strconv.Atoi(p); err == nil {
				switch {
				case (n >= 30 && n <= 37) || (n >= 90 && n <= 97):
					fg = false
				case (n >= 40 && n <= 47) || (n >= 100 && n <= 107):
					bg = false
				}
			}
		}
	}
	return fg, bg
}
