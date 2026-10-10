package themes

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadThemesFromDirectoryLoadsExampleThemes(t *testing.T) {
	t.Parallel()

	themesDir := filepath.Join("..", "..", "examples", "themes")
	themes, err := LoadThemesFromDirectory(themesDir)
	if err != nil {
		t.Fatalf("LoadThemesFromDirectory returned error: %v", err)
	}
	if len(themes) != 1 || themes[0].Name != "custom" {
		t.Fatalf("expected the single custom example theme, got %d", len(themes))
	}
	custom := themes[0]
	if custom.Description == "" {
		t.Fatalf("expected example theme description to be preserved")
	}
	// Palette entries may reference named colors.
	if got := Hex(custom.Palette.Base); got != "#1B1411" {
		t.Fatalf("palette base = %s, want #1B1411 (resolved from colors.espresso)", got)
	}
	if got := Hex(custom.BaseBg); got != "#1B1411" {
		t.Fatalf("BaseBg = %s, want it derived from the palette", got)
	}
}

func TestBuiltinThemesLoadWithDistinctPalettes(t *testing.T) {
	t.Parallel()

	themes := BuiltinThemes()
	if len(themes) < 10 {
		t.Fatalf("expected default + 9 built-in themes, got %d", len(themes))
	}
	if themes[0].Name != "default" {
		t.Fatalf("default theme must come first, got %q", themes[0].Name)
	}
	seen := map[string]bool{}
	for _, theme := range themes {
		if seen[theme.Name] {
			t.Fatalf("duplicate built-in theme %q", theme.Name)
		}
		seen[theme.Name] = true
		if theme.Name != "default" && Hex(theme.BaseBg) == Hex(DefaultPalette().Base) {
			t.Errorf("theme %q still uses the default canvas color", theme.Name)
		}
		if Hex(theme.SurfaceBg) != Hex(theme.Palette.Overlay) {
			t.Errorf("theme %q: dialogs must sit on the palette overlay", theme.Name)
		}
	}
	if !seen["light"] {
		t.Fatal("expected a built-in light theme")
	}
}

func TestCustomThemeReplacesBuiltinOfSameName(t *testing.T) {
	t.Parallel()

	custom := NewThemeFromPalette("light", "mine", DefaultPalette())
	merged := mergeThemes(BuiltinThemes(), []Theme{custom})
	count := 0
	for _, theme := range merged {
		if theme.Name == "light" {
			count++
			if theme.Description != "mine" {
				t.Fatalf("custom light theme should replace the built-in one")
			}
		}
	}
	if count != 1 {
		t.Fatalf("expected exactly one light theme, got %d", count)
	}
}

func TestLoadThemeFromYAMLRejectsUnknownPaletteKey(t *testing.T) {
	t.Parallel()

	path := writeThemeFixture(t, `name: "bad-palette"
palette:
  backgroundz: "#ffffff"
`)
	if _, err := LoadThemeFromYAML(path); err == nil || !strings.Contains(err.Error(), "unknown palette key") {
		t.Fatalf("expected unknown palette key error, got %v", err)
	}
}

func TestLoadThemeFromYAMLRejectsUnknownStyleKey(t *testing.T) {
	t.Parallel()

	path := writeThemeFixture(t, `name: "invalid"
styles:
  not_a_real_style:
    foreground: "#ffffff"
`)

	_, err := LoadThemeFromYAML(path)
	if err == nil {
		t.Fatal("expected error for unknown style key")
	}
	if !strings.Contains(err.Error(), `unknown style key "not_a_real_style"`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoadThemeFromYAMLAcceptsToolCallStyleKeys(t *testing.T) {
	t.Parallel()

	path := writeThemeFixture(t, `name: "tool-call"
styles:
  tool_call_label:
    foreground: "#ffffff"
  tool_call_input:
    foreground: "#888888"
`)

	if _, err := LoadThemeFromYAML(path); err != nil {
		t.Fatalf("expected tool call style keys to be accepted, got error: %v", err)
	}
}

func TestLoadThemeFromYAMLRejectsInvalidPadding(t *testing.T) {
	t.Parallel()

	path := writeThemeFixture(t, `name: "invalid-padding"
styles:
  title:
    padding: "1,2,3"
`)

	_, err := LoadThemeFromYAML(path)
	if err == nil {
		t.Fatal("expected error for invalid padding")
	}
	if !strings.Contains(err.Error(), `invalid padding`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoadThemesFromDirectorySkipsDuplicateNames(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeThemeFile(t, filepath.Join(dir, "first.yaml"), `name: "shared"
styles:
  title:
    foreground: "#ffffff"
`)
	writeThemeFile(t, filepath.Join(dir, "second.yaml"), `name: "shared"
styles:
  title:
    foreground: "#000000"
`)

	themes, err := LoadThemesFromDirectory(dir)
	if err != nil {
		t.Fatalf("LoadThemesFromDirectory returned error: %v", err)
	}

	if len(themes) != 1 {
		t.Fatalf("expected duplicate theme names to be skipped, got %d themes", len(themes))
	}
}

func writeThemeFixture(t *testing.T, contents string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "theme.yaml")
	writeThemeFile(t, path, contents)
	return path
}

func writeThemeFile(t *testing.T, path string, contents string) {
	t.Helper()

	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("failed to write theme fixture: %v", err)
	}
}
