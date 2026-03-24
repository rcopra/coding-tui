package ui

import (
	"math"
	"strings"

	"charm.land/glamour/v2/ansi"
	"charm.land/glamour/v2/styles"
	"charm.land/lipgloss/v2"
	xansi "github.com/charmbracelet/x/ansi"
)

// Color palette: Tokyo Night inspired, matching OpenCode's aesthetic.
// High contrast white text on true black. Reads like a document.
var (
	black  = lipgloss.Color("#000000") // true black background
	white  = lipgloss.Color("#c0caf5") // body text (tokyo night fg)
	dim    = lipgloss.Color("#565f89") // muted text, comments, chrome
	faint  = lipgloss.Color("#24283b") // scrollbar track, very subtle
	accent = lipgloss.Color("#7aa2f7") // bright blue — headings, primary
	gold   = lipgloss.Color("#e0af68") // golden — section headings, warnings
	green  = lipgloss.Color("#9ece6a") // success, easy, strings
	yellow = lipgloss.Color("#e0af68") // medium difficulty
	red    = lipgloss.Color("#f7768e") // errors, hard, inline code
	peach  = lipgloss.Color("#ff9e64") // status messages, numbers
	purple = lipgloss.Color("#bb9af7") // keywords, accent2
)

func stringPtr(s string) *string { return &s }
func uintPtr(u uint) *uint       { return &u }
func boolPtr(b bool) *bool       { return &b }

// exercismGlamourStyle returns a Tokyo Night glamour style
// matching OpenCode's rendering aesthetic.
func exercismGlamourStyle() ansi.StyleConfig {
	s := styles.DarkStyleConfig

	// Document: clean white text, tight margin
	s.Document.StylePrimitive.Color = stringPtr("#c0caf5")
	s.Document.Margin = uintPtr(1)

	// Headings: bright blue like OpenCode
	s.Heading.StylePrimitive.Color = stringPtr("#7aa2f7")
	s.Heading.StylePrimitive.Bold = boolPtr(true)

	s.H1.StylePrimitive.Color = stringPtr("#7aa2f7")
	s.H1.StylePrimitive.BackgroundColor = nil
	s.H1.StylePrimitive.Bold = boolPtr(true)
	s.H1.StylePrimitive.Prefix = " "
	s.H1.StylePrimitive.Suffix = " "

	s.H2.StylePrimitive.Color = stringPtr("#e0af68")
	s.H3.StylePrimitive.Color = stringPtr("#e0af68")
	s.H4.StylePrimitive.Color = stringPtr("#7aa2f7")
	s.H5.StylePrimitive.Color = stringPtr("#7aa2f7")
	s.H6.StylePrimitive.Color = stringPtr("#565f89")

	// Inline code: tokyo night red on dark surface
	s.Code = ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Prefix:          "\u00a0",
			Suffix:          "\u00a0",
			Color:           stringPtr("#f7768e"),
			BackgroundColor: stringPtr("#1a1b26"),
		},
	}

	// Code blocks: Tokyo Night syntax highlighting
	s.CodeBlock = ansi.StyleCodeBlock{
		StyleBlock: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color: stringPtr("#c0caf5"),
			},
			Margin: uintPtr(1),
		},
		Chroma: &ansi.Chroma{
			Text:    ansi.StylePrimitive{Color: stringPtr("#c0caf5")},
			Error:   ansi.StylePrimitive{Color: stringPtr("#f7768e")},
			Comment: ansi.StylePrimitive{Color: stringPtr("#565f89")},
			CommentPreproc: ansi.StylePrimitive{Color: stringPtr("#ff9e64")},
			Keyword:          ansi.StylePrimitive{Color: stringPtr("#bb9af7")},
			KeywordReserved:  ansi.StylePrimitive{Color: stringPtr("#bb9af7")},
			KeywordNamespace: ansi.StylePrimitive{Color: stringPtr("#bb9af7")},
			KeywordType:      ansi.StylePrimitive{Color: stringPtr("#2ac3de")},
			Operator:         ansi.StylePrimitive{Color: stringPtr("#89ddff")},
			Punctuation:      ansi.StylePrimitive{Color: stringPtr("#c0caf5")},
			Name:             ansi.StylePrimitive{Color: stringPtr("#c0caf5")},
			NameBuiltin:      ansi.StylePrimitive{Color: stringPtr("#ff9e64")},
			NameTag:          ansi.StylePrimitive{Color: stringPtr("#f7768e")},
			NameAttribute:    ansi.StylePrimitive{Color: stringPtr("#e0af68")},
			NameClass:        ansi.StylePrimitive{Color: stringPtr("#e0af68"), Bold: boolPtr(true)},
			NameDecorator:    ansi.StylePrimitive{Color: stringPtr("#e0af68")},
			NameFunction:     ansi.StylePrimitive{Color: stringPtr("#7aa2f7")},
			NameConstant:     ansi.StylePrimitive{Color: stringPtr("#ff9e64")},
			NameException:    ansi.StylePrimitive{Color: stringPtr("#f7768e")},
			LiteralNumber:    ansi.StylePrimitive{Color: stringPtr("#ff9e64")},
			LiteralString:    ansi.StylePrimitive{Color: stringPtr("#9ece6a")},
			LiteralStringEscape: ansi.StylePrimitive{Color: stringPtr("#89ddff")},
			GenericDeleted:   ansi.StylePrimitive{Color: stringPtr("#f7768e")},
			GenericEmph:      ansi.StylePrimitive{Italic: boolPtr(true)},
			GenericInserted:  ansi.StylePrimitive{Color: stringPtr("#9ece6a")},
			GenericStrong:    ansi.StylePrimitive{Bold: boolPtr(true)},
			GenericSubheading: ansi.StylePrimitive{Color: stringPtr("#7aa2f7")},
		},
	}

	// Links: blue
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

	// Horizontal rule
	s.HorizontalRule = ansi.StylePrimitive{
		Color:  stringPtr("#e0af68"),
		Format: "\n───\n",
	}

	// Block quotes
	s.BlockQuote = ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Color: stringPtr("#565f89"),
		},
		Indent:      uintPtr(1),
		IndentToken: stringPtr("│ "),
	}

	return s
}

// renderScrollbar renders a thin scrollbar track alongside content.
func renderScrollbar(content string, viewportHeight, totalLines, offset int) string {
	if totalLines <= viewportHeight || viewportHeight <= 0 {
		return content
	}

	lines := strings.Split(content, "\n")

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

// fillBackground pads every line of content to the given width with
// a true black background, eliminating any gaps/banding.
func fillBackground(content string, width int) string {
	bgStyle := lipgloss.NewStyle().Background(black)
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		visible := xansi.StringWidth(line)
		if visible < width {
			line += strings.Repeat(" ", width-visible)
		}
		lines[i] = bgStyle.Render(line)
	}
	return strings.Join(lines, "\n")
}
