// Package markdown builds the app's themed glamour renderer, so every markdown
// surface (chat conversation, recipe detail, cooking chat) shares one palette
// derived from the active theme rather than glamour's environment default.
package markdown

import (
	"strings"

	"charm.land/glamour/v2"
	"charm.land/glamour/v2/ansi"
	themes "github.com/GarroshIcecream/yummy/internal/themes"
)

// NewThemedRenderer creates a glamour renderer whose StyleConfig is derived
// from the theme's palette, so rendered markdown matches the rest of the UI
// under any theme. A nil theme uses the default theme. Extra options (e.g.
// glamour.WithEmoji) are appended after the themed defaults.
//
// Glamour leaves the background unset (it emits a default-background reset
// after every span); callers sit the output on their surface and the app-wide
// canvas fill re-establishes the theme background behind it.
func NewThemedRenderer(theme *themes.Theme, wordWrap int, extra ...glamour.TermRendererOption) (*glamour.TermRenderer, error) {
	if theme == nil {
		def := themes.NewDefaultTheme()
		theme = &def
	}
	p := theme.Palette
	hex := themes.Hex
	s := func(c string) *string { return &c }
	b := boolPtr
	u := uintPtr

	text := hex(p.Text)
	muted := hex(p.Muted)
	subtle := hex(p.Subtle)
	border := hex(p.Border)
	primary := hex(p.Primary)
	secondary := hex(p.Secondary)
	accent := hex(p.Accent)
	success := hex(p.Success)
	warning := hex(p.Warning)
	errorColor := hex(p.Error)

	style := ansi.StyleConfig{
		Document: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				BlockPrefix: "\n",
				BlockSuffix: "\n",
				Color:       s(text),
			},
			Margin: u(2),
		},
		BlockQuote: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color: s(muted),
			},
			Indent:      u(1),
			IndentToken: s("│ "),
		},
		List: ansi.StyleList{
			LevelIndent: 2,
		},
		// Headings are told apart by color and weight rather than literal
		// "#" markers: H1/H2 in the primary accent, H3 in the secondary.
		Heading: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				BlockSuffix: "\n",
				Color:       s(primary),
				Bold:        b(true),
			},
		},
		H1: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color: s(primary),
				Bold:  b(true),
				Upper: b(true),
			},
		},
		H2: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color: s(primary),
				Bold:  b(true),
			},
		},
		H3: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color: s(secondary),
				Bold:  b(true),
			},
		},
		H4: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color: s(text),
				Bold:  b(true),
			},
		},
		H5: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color: s(muted),
				Bold:  b(true),
			},
		},
		H6: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color: s(subtle),
				Bold:  b(false),
			},
		},
		Strikethrough: ansi.StylePrimitive{
			CrossedOut: b(true),
		},
		Emph: ansi.StylePrimitive{
			Italic: b(true),
		},
		Strong: ansi.StylePrimitive{
			Bold: b(true),
		},
		HorizontalRule: ansi.StylePrimitive{
			Color:  s(border),
			Format: "\n" + strings.Repeat("─", max(wordWrap-4, 8)) + "\n",
		},
		Item: ansi.StylePrimitive{
			BlockPrefix: "> ",
		},
		Enumeration: ansi.StylePrimitive{
			BlockPrefix: ". ",
		},
		Task: ansi.StyleTask{
			StylePrimitive: ansi.StylePrimitive{},
			Ticked:         "[✓] ",
			Unticked:       "[ ] ",
		},
		Link: ansi.StylePrimitive{
			Color:     s(secondary),
			Underline: b(true),
		},
		LinkText: ansi.StylePrimitive{
			Color: s(secondary),
			Bold:  b(true),
		},
		Image: ansi.StylePrimitive{
			Color:     s(secondary),
			Underline: b(true),
		},
		ImageText: ansi.StylePrimitive{
			Color:  s(muted),
			Format: "Image: {{.text}} →",
		},
		Code: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Prefix:          "\u00a0",
				Suffix:          "\u00a0",
				Color:           s(accent),
				BackgroundColor: s(hex(p.Selection)),
			},
		},
		CodeBlock: ansi.StyleCodeBlock{
			StyleBlock: ansi.StyleBlock{
				StylePrimitive: ansi.StylePrimitive{
					Color: s(muted),
				},
				Margin: u(2),
			},
			Chroma: &ansi.Chroma{
				Text:    ansi.StylePrimitive{Color: s(text)},
				Comment: ansi.StylePrimitive{Color: s(subtle)},
				CommentPreproc: ansi.StylePrimitive{
					Color: s(warning),
				},
				Keyword: ansi.StylePrimitive{
					Color: s(primary),
				},
				KeywordReserved: ansi.StylePrimitive{
					Color: s(secondary),
				},
				KeywordNamespace: ansi.StylePrimitive{
					Color: s(errorColor),
				},
				KeywordType: ansi.StylePrimitive{
					Color: s(primary),
				},
				Operator: ansi.StylePrimitive{
					Color: s(errorColor),
				},
				Punctuation: ansi.StylePrimitive{
					Color: s(muted),
				},
				Name: ansi.StylePrimitive{
					Color: s(text),
				},
				NameBuiltin: ansi.StylePrimitive{
					Color: s(secondary),
				},
				NameTag: ansi.StylePrimitive{
					Color: s(secondary),
				},
				NameAttribute: ansi.StylePrimitive{
					Color: s(secondary),
				},
				NameClass: ansi.StylePrimitive{
					Color:     s(text),
					Underline: b(true),
					Bold:      b(true),
				},
				NameDecorator: ansi.StylePrimitive{
					Color: s(accent),
				},
				NameFunction: ansi.StylePrimitive{
					Color: s(success),
				},
				LiteralNumber: ansi.StylePrimitive{
					Color: s(secondary),
				},
				LiteralString: ansi.StylePrimitive{
					Color: s(accent),
				},
				LiteralStringEscape: ansi.StylePrimitive{
					Color: s(success),
				},
				GenericDeleted: ansi.StylePrimitive{
					Color: s(errorColor),
				},
				GenericEmph: ansi.StylePrimitive{
					Italic: b(true),
				},
				GenericInserted: ansi.StylePrimitive{
					Color: s(success),
				},
				GenericStrong: ansi.StylePrimitive{
					Bold: b(true),
				},
				GenericSubheading: ansi.StylePrimitive{
					Color: s(subtle),
				},
				// No Background — inherits from the viewport/app bg.
			},
		},
		Table: ansi.StyleTable{
			StyleBlock: ansi.StyleBlock{
				StylePrimitive: ansi.StylePrimitive{},
			},
		},
		DefinitionDescription: ansi.StylePrimitive{
			BlockPrefix: "\n🠶 ",
		},
	}

	opts := []glamour.TermRendererOption{
		glamour.WithStyles(style),
		glamour.WithWordWrap(wordWrap),
	}
	opts = append(opts, extra...)
	return glamour.NewTermRenderer(opts...)
}

//go:fix inline
func boolPtr(b bool) *bool { return new(b) }

//go:fix inline
func uintPtr(u uint) *uint { return new(u) }
