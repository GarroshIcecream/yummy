package themes

import (
	"fmt"
	"image/color"
	"strings"

	lipgloss "charm.land/lipgloss/v2"
)

// Palette is the small set of semantic colors every theme is built from.
// NewThemeFromPalette derives every style in the app from these, so a theme
// only has to pick ~15 colors to restyle the whole UI consistently.
//
// Surfaces, from back to front:
//
//	Base      the app canvas (every view sits on it)
//	Surface   slightly raised chrome on the canvas: status bar, selected menu item
//	Overlay   popover dialogs floating above the canvas
//	Selection highlighted row inside lists and dialogs
type Palette struct {
	Base      color.Color
	Surface   color.Color
	Overlay   color.Color
	Selection color.Color
	Border    color.Color

	Text   color.Color // primary text
	Muted  color.Color // secondary text (descriptions)
	Subtle color.Color // hints, help, placeholders

	Primary   color.Color // main accent: titles, focus, selection
	Secondary color.Color // second accent: assistant, links
	Accent    color.Color // warm highlight: amounts, stars, timers
	Success   color.Color
	Warning   color.Color
	Error     color.Color

	OnPrimary color.Color // text drawn on a Primary background
}

// DefaultPalette is the built-in dark palette.
func DefaultPalette() Palette {
	return Palette{
		Base:      lipgloss.Color("#0F1115"),
		Surface:   lipgloss.Color("#171A21"),
		Overlay:   lipgloss.Color("#1D2330"),
		Selection: lipgloss.Color("#2A3448"),
		Border:    lipgloss.Color("#2B3445"),

		Text:   lipgloss.Color("#E6EAF2"),
		Muted:  lipgloss.Color("#8B93A7"),
		Subtle: lipgloss.Color("#626B7F"),

		Primary:   lipgloss.Color("#4A9EFF"),
		Secondary: lipgloss.Color("#4ECDC4"),
		Accent:    lipgloss.Color("#F0C674"),
		Success:   lipgloss.Color("#3FB950"),
		Warning:   lipgloss.Color("#FF9F43"),
		Error:     lipgloss.Color("#FF6B6B"),

		OnPrimary: lipgloss.Color("#0B0E14"),
	}
}

// paletteKeys maps YAML palette keys to the palette field they set.
func (p *Palette) fields() map[string]*color.Color {
	return map[string]*color.Color{
		"base":       &p.Base,
		"surface":    &p.Surface,
		"overlay":    &p.Overlay,
		"selection":  &p.Selection,
		"border":     &p.Border,
		"text":       &p.Text,
		"muted":      &p.Muted,
		"subtle":     &p.Subtle,
		"primary":    &p.Primary,
		"secondary":  &p.Secondary,
		"accent":     &p.Accent,
		"success":    &p.Success,
		"warning":    &p.Warning,
		"error":      &p.Error,
		"on_primary": &p.OnPrimary,
	}
}

// PaletteKeys returns the YAML keys accepted in a theme's `palette:` section.
func PaletteKeys() []string {
	return []string{
		"base", "surface", "overlay", "selection", "border",
		"text", "muted", "subtle",
		"primary", "secondary", "accent", "success", "warning", "error",
		"on_primary",
	}
}

// apply overrides palette entries from a YAML `palette:` map. Values are hex
// colors or names from the theme's `colors:` map (resolved by resolve).
func (p *Palette) apply(entries map[string]string, resolve func(string) string) error {
	fields := p.fields()
	for key, value := range entries {
		field, ok := fields[strings.ToLower(key)]
		if !ok {
			return fmt.Errorf("unknown palette key %q (expected one of: %s)", key, strings.Join(PaletteKeys(), ", "))
		}
		hex := resolve(value)
		if !isHexColor(hex) {
			return fmt.Errorf("palette %q: %q is not a hex color", key, value)
		}
		*field = lipgloss.Color(hex)
	}
	return nil
}

func isHexColor(s string) bool {
	if !strings.HasPrefix(s, "#") || (len(s) != 4 && len(s) != 7) {
		return false
	}
	for _, r := range s[1:] {
		if !strings.ContainsRune("0123456789abcdefABCDEF", r) {
			return false
		}
	}
	return true
}

// Hex renders a color as #RRGGBB (used by renderers that take color strings,
// such as glamour).
func Hex(c color.Color) string {
	if c == nil {
		return ""
	}
	r, g, b, _ := c.RGBA()
	return fmt.Sprintf("#%02X%02X%02X", r>>8, g>>8, b>>8)
}
