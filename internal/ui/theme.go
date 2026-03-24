package ui

import (
	"charm.land/glamour/v2/ansi"
	"charm.land/glamour/v2/styles"
	"charm.land/lipgloss/v2"
)

// Color palette: gruvbox warmth + catppuccin purple accents.
// Designed to sit alongside gruvbox in a split tmux pane
// while having its own identity.
var (
	// Catppuccin Mocha accents
	mauve    = lipgloss.Color("#cba6f7") // primary accent — the purple
	lavender = lipgloss.Color("#b4befe") // secondary accent
	peach    = lipgloss.Color("#fab387") // warnings, status messages
	green    = lipgloss.Color("#a6e3a1") // success, downloaded, passed
	red      = lipgloss.Color("#f38ba8") // errors, hard difficulty
	yellow   = lipgloss.Color("#f9e2af") // medium difficulty, inline code
	sky      = lipgloss.Color("#89dceb") // links, info
	text     = lipgloss.Color("#cdd6f4") // body text
	subtext  = lipgloss.Color("#a6adc8") // muted text
	overlay  = lipgloss.Color("#6c7086") // subtle/dim text
	surface  = lipgloss.Color("#313244") // code block backgrounds
)

func stringPtr(s string) *string { return &s }
func uintPtr(u uint) *uint       { return &u }
func boolPtr(b bool) *bool       { return &b }

// exercismGlamourStyle returns a custom glamour style tuned for
// readability in a terminal alongside gruvbox.
func exercismGlamourStyle() ansi.StyleConfig {
	s := styles.DarkStyleConfig

	// Document: tighter margin, warm text
	s.Document.StylePrimitive.Color = stringPtr("#cdd6f4")
	s.Document.Margin = uintPtr(1)

	// Headings: catppuccin purple accent
	s.Heading.StylePrimitive.Color = stringPtr("#cba6f7")
	s.Heading.StylePrimitive.Bold = boolPtr(true)

	s.H1.StylePrimitive.Color = stringPtr("#1e1e2e")
	s.H1.StylePrimitive.BackgroundColor = stringPtr("#cba6f7")
	s.H1.StylePrimitive.Bold = boolPtr(true)

	s.H2.StylePrimitive.Color = stringPtr("#cba6f7")
	s.H3.StylePrimitive.Color = stringPtr("#b4befe")
	s.H4.StylePrimitive.Color = stringPtr("#b4befe")
	s.H5.StylePrimitive.Color = stringPtr("#b4befe")
	s.H6.StylePrimitive.Color = stringPtr("#6c7086")

	// Inline code: warm yellow on dark surface — readable, not jarring
	s.Code = ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Prefix:          "\u00a0",
			Suffix:          "\u00a0",
			Color:           stringPtr("#f9e2af"),
			BackgroundColor: stringPtr("#313244"),
		},
	}

	// Code blocks: catppuccin-flavored syntax highlighting
	s.CodeBlock = ansi.StyleCodeBlock{
		StyleBlock: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color: stringPtr("#cdd6f4"),
			},
			Margin: uintPtr(1),
		},
		Chroma: &ansi.Chroma{
			Text:    ansi.StylePrimitive{Color: stringPtr("#cdd6f4")},
			Error:   ansi.StylePrimitive{Color: stringPtr("#f38ba8")},
			Comment: ansi.StylePrimitive{Color: stringPtr("#6c7086")},
			CommentPreproc: ansi.StylePrimitive{Color: stringPtr("#fab387")},
			Keyword:          ansi.StylePrimitive{Color: stringPtr("#cba6f7")},
			KeywordReserved:  ansi.StylePrimitive{Color: stringPtr("#f38ba8")},
			KeywordNamespace: ansi.StylePrimitive{Color: stringPtr("#89dceb")},
			KeywordType:      ansi.StylePrimitive{Color: stringPtr("#f9e2af")},
			Operator:         ansi.StylePrimitive{Color: stringPtr("#89dceb")},
			Punctuation:      ansi.StylePrimitive{Color: stringPtr("#a6adc8")},
			Name:             ansi.StylePrimitive{Color: stringPtr("#cdd6f4")},
			NameBuiltin:      ansi.StylePrimitive{Color: stringPtr("#fab387")},
			NameTag:          ansi.StylePrimitive{Color: stringPtr("#cba6f7")},
			NameAttribute:    ansi.StylePrimitive{Color: stringPtr("#f9e2af")},
			NameClass:        ansi.StylePrimitive{Color: stringPtr("#f9e2af"), Bold: boolPtr(true)},
			NameDecorator:    ansi.StylePrimitive{Color: stringPtr("#f9e2af")},
			NameFunction:     ansi.StylePrimitive{Color: stringPtr("#89b4fa")},
			NameConstant:     ansi.StylePrimitive{Color: stringPtr("#fab387")},
			NameException:    ansi.StylePrimitive{Color: stringPtr("#f38ba8")},
			LiteralNumber:    ansi.StylePrimitive{Color: stringPtr("#fab387")},
			LiteralString:    ansi.StylePrimitive{Color: stringPtr("#a6e3a1")},
			LiteralStringEscape: ansi.StylePrimitive{Color: stringPtr("#f2cdcd")},
			GenericDeleted:   ansi.StylePrimitive{Color: stringPtr("#f38ba8")},
			GenericEmph:      ansi.StylePrimitive{Italic: boolPtr(true)},
			GenericInserted:  ansi.StylePrimitive{Color: stringPtr("#a6e3a1")},
			GenericStrong:    ansi.StylePrimitive{Bold: boolPtr(true)},
			GenericSubheading: ansi.StylePrimitive{Color: stringPtr("#89dceb")},
		},
	}

	// Links: sky blue
	s.Link = ansi.StylePrimitive{
		Color:     stringPtr("#89dceb"),
		Underline: boolPtr(true),
	}
	s.LinkText = ansi.StylePrimitive{
		Color: stringPtr("#89dceb"),
		Bold:  boolPtr(true),
	}

	// Strong/emphasis
	s.Strong = ansi.StylePrimitive{Bold: boolPtr(true)}
	s.Emph = ansi.StylePrimitive{Italic: boolPtr(true)}

	// Horizontal rule
	s.HorizontalRule = ansi.StylePrimitive{
		Color:  stringPtr("#6c7086"),
		Format: "\n────────\n",
	}

	// Block quotes: lavender indent
	s.BlockQuote = ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Color: stringPtr("#a6adc8"),
		},
		Indent:      uintPtr(1),
		IndentToken: stringPtr("│ "),
	}

	return s
}
