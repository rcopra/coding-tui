package ui

import (
	"math"
	"strings"

	"charm.land/glamour/v2/ansi"
	"charm.land/glamour/v2/styles"
	"charm.land/lipgloss/v2"
)

// Color palette: high contrast, OpenCode-inspired.
// White text on dark background. Reads like a document, not an editor.
// Minimal color — only used for structure and emphasis.
var (
	// Core
	white   = lipgloss.Color("#e4e4e4") // body text — bright but not blinding
	dim     = lipgloss.Color("#6e6e6e") // muted text, comments, chrome
	faint   = lipgloss.Color("#3a3a3a") // very subtle, scrollbar track
	accent  = lipgloss.Color("#d4a052") // warm gold — headings, primary accent
	accent2 = lipgloss.Color("#7aa2f7") // cool blue — links, secondary
	green   = lipgloss.Color("#73c936") // success, easy, strings
	yellow  = lipgloss.Color("#d4a052") // medium difficulty, warnings
	red     = lipgloss.Color("#d75f5f") // errors, hard difficulty
	peach   = lipgloss.Color("#d7875f") // status messages, numbers
	coral   = lipgloss.Color("#d78787") // inline code text
)

func stringPtr(s string) *string { return &s }
func uintPtr(u uint) *uint       { return &u }
func boolPtr(b bool) *bool       { return &b }

// exercismGlamourStyle returns a high-contrast glamour style
// inspired by OpenCode — white text, warm gold headings,
// quiet syntax highlighting. Designed for reading.
func exercismGlamourStyle() ansi.StyleConfig {
	s := styles.DarkStyleConfig

	// Document: clean white text, tight margin
	s.Document.StylePrimitive.Color = stringPtr("#e4e4e4")
	s.Document.Margin = uintPtr(1)

	// Headings: warm gold accent — the only real color pop
	s.Heading.StylePrimitive.Color = stringPtr("#d4a052")
	s.Heading.StylePrimitive.Bold = boolPtr(true)

	s.H1.StylePrimitive.Color = stringPtr("#d4a052")
	s.H1.StylePrimitive.BackgroundColor = nil
	s.H1.StylePrimitive.Bold = boolPtr(true)
	s.H1.StylePrimitive.Prefix = " "
	s.H1.StylePrimitive.Suffix = " "

	s.H2.StylePrimitive.Color = stringPtr("#d4a052")
	s.H3.StylePrimitive.Color = stringPtr("#d4a052")
	s.H4.StylePrimitive.Color = stringPtr("#d7af87")
	s.H5.StylePrimitive.Color = stringPtr("#d7af87")
	s.H6.StylePrimitive.Color = stringPtr("#6e6e6e")

	// Inline code: soft coral on subtle background
	s.Code = ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Prefix:          "\u00a0",
			Suffix:          "\u00a0",
			Color:           stringPtr("#d78787"),
			BackgroundColor: stringPtr("#262626"),
		},
	}

	// Code blocks: restrained syntax highlighting
	s.CodeBlock = ansi.StyleCodeBlock{
		StyleBlock: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color: stringPtr("#c0c0c0"),
			},
			Margin: uintPtr(1),
		},
		Chroma: &ansi.Chroma{
			Text:    ansi.StylePrimitive{Color: stringPtr("#c0c0c0")},
			Error:   ansi.StylePrimitive{Color: stringPtr("#d75f5f")},
			Comment: ansi.StylePrimitive{Color: stringPtr("#5f5f5f")},
			CommentPreproc: ansi.StylePrimitive{Color: stringPtr("#d7875f")},
			Keyword:          ansi.StylePrimitive{Color: stringPtr("#af87d7")},
			KeywordReserved:  ansi.StylePrimitive{Color: stringPtr("#af87d7")},
			KeywordNamespace: ansi.StylePrimitive{Color: stringPtr("#af87d7")},
			KeywordType:      ansi.StylePrimitive{Color: stringPtr("#d4a052")},
			Operator:         ansi.StylePrimitive{Color: stringPtr("#87afaf")},
			Punctuation:      ansi.StylePrimitive{Color: stringPtr("#808080")},
			Name:             ansi.StylePrimitive{Color: stringPtr("#c0c0c0")},
			NameBuiltin:      ansi.StylePrimitive{Color: stringPtr("#d7875f")},
			NameTag:          ansi.StylePrimitive{Color: stringPtr("#af87d7")},
			NameAttribute:    ansi.StylePrimitive{Color: stringPtr("#d4a052")},
			NameClass:        ansi.StylePrimitive{Color: stringPtr("#d4a052"), Bold: boolPtr(true)},
			NameDecorator:    ansi.StylePrimitive{Color: stringPtr("#d4a052")},
			NameFunction:     ansi.StylePrimitive{Color: stringPtr("#7aa2f7")},
			NameConstant:     ansi.StylePrimitive{Color: stringPtr("#d7875f")},
			NameException:    ansi.StylePrimitive{Color: stringPtr("#d75f5f")},
			LiteralNumber:    ansi.StylePrimitive{Color: stringPtr("#d7875f")},
			LiteralString:    ansi.StylePrimitive{Color: stringPtr("#73c936")},
			LiteralStringEscape: ansi.StylePrimitive{Color: stringPtr("#87afaf")},
			GenericDeleted:   ansi.StylePrimitive{Color: stringPtr("#d75f5f")},
			GenericEmph:      ansi.StylePrimitive{Italic: boolPtr(true)},
			GenericInserted:  ansi.StylePrimitive{Color: stringPtr("#73c936")},
			GenericStrong:    ansi.StylePrimitive{Bold: boolPtr(true)},
			GenericSubheading: ansi.StylePrimitive{Color: stringPtr("#7aa2f7")},
		},
	}

	// Links: cool blue — the only other color
	s.Link = ansi.StylePrimitive{
		Color:     stringPtr("#7aa2f7"),
		Underline: boolPtr(true),
	}
	s.LinkText = ansi.StylePrimitive{
		Color: stringPtr("#7aa2f7"),
	}

	// Strong/emphasis
	s.Strong = ansi.StylePrimitive{Bold: boolPtr(true)}
	s.Emph = ansi.StylePrimitive{Italic: boolPtr(true)}

	// Horizontal rule: quiet
	s.HorizontalRule = ansi.StylePrimitive{
		Color:  stringPtr("#3a3a3a"),
		Format: "\n───\n",
	}

	// Block quotes
	s.BlockQuote = ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Color: stringPtr("#808080"),
		},
		Indent:      uintPtr(1),
		IndentToken: stringPtr("│ "),
	}

	return s
}

// renderScrollbar renders a thin scrollbar track alongside content.
// Returns the content with a scrollbar column appended to the right.
func renderScrollbar(content string, viewportHeight, totalLines, offset int) string {
	if totalLines <= viewportHeight || viewportHeight <= 0 {
		return content
	}

	lines := strings.Split(content, "\n")

	// Calculate thumb position and size
	thumbSize := int(math.Max(1, math.Round(float64(viewportHeight)*float64(viewportHeight)/float64(totalLines))))
	thumbStart := int(math.Round(float64(offset) * float64(viewportHeight-thumbSize) / float64(totalLines-viewportHeight)))

	trackChar := lipgloss.NewStyle().Foreground(faint).Render("│")
	thumbChar := lipgloss.NewStyle().Foreground(dim).Render("┃")

	for i := 0; i < len(lines) && i < viewportHeight; i++ {
		char := trackChar
		if i >= thumbStart && i < thumbStart+thumbSize {
			char = thumbChar
		}
		lines[i] = lines[i] + " " + char
	}

	return strings.Join(lines, "\n")
}
