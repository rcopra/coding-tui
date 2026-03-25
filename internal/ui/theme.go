package ui

import (
	"math"
	"strings"

	"charm.land/glamour/v2/ansi"
	"charm.land/glamour/v2/styles"
	"charm.land/lipgloss/v2"
	xansi "github.com/charmbracelet/x/ansi"
)

// Color palette: OpenCode-inspired.
// True black bg, bright white text, purple accent headings.
var (
	black  = lipgloss.Color("#0a0a0a") // background
	white  = lipgloss.Color("#eeeeee") // body text
	dim    = lipgloss.Color("#808080") // muted text
	faint  = lipgloss.Color("#1e1e1e") // scrollbar track, surfaces
	accent = lipgloss.Color("#9d7cd8") // purple — headings, primary
	blue   = lipgloss.Color("#5c9cf5") // secondary, links
	green  = lipgloss.Color("#7fd88f") // success, easy, strings
	yellow = lipgloss.Color("#e5c07b") // medium difficulty, types
	red    = lipgloss.Color("#e06c75") // errors, hard, inline code
	peach  = lipgloss.Color("#f5a742") // status messages, numbers
	purple = lipgloss.Color("#9d7cd8") // keywords
	cyan   = lipgloss.Color("#56b6c2") // operators, info
)

func stringPtr(s string) *string { return &s }
func uintPtr(u uint) *uint       { return &u }
func boolPtr(b bool) *bool       { return &b }

// exercismGlamourStyle returns a glamour style matching OpenCode's aesthetic.
func exercismGlamourStyle() ansi.StyleConfig {
	s := styles.DarkStyleConfig

	// Document: bright white text, clean margin
	s.Document.StylePrimitive.Color = stringPtr("#eeeeee")
	s.Document.Margin = uintPtr(2)

	// Headings: purple accent like OpenCode — no ## prefix markers
	s.Heading.StylePrimitive.Color = stringPtr("#9d7cd8")
	s.Heading.StylePrimitive.Bold = boolPtr(true)

	s.H1.StylePrimitive.Color = stringPtr("#5c9cf5")
	s.H1.StylePrimitive.BackgroundColor = nil
	s.H1.StylePrimitive.Bold = boolPtr(true)
	s.H1.StylePrimitive.Prefix = ""
	s.H1.StylePrimitive.Suffix = ""

	// Remove the typewriter ## prefix markers — cleaner like OpenCode
	s.H2.StylePrimitive.Color = stringPtr("#e5c07b")
	s.H2.StylePrimitive.Prefix = ""

	s.H3.StylePrimitive.Color = stringPtr("#e5c07b")
	s.H3.StylePrimitive.Prefix = ""

	s.H4.StylePrimitive.Color = stringPtr("#9d7cd8")
	s.H4.StylePrimitive.Prefix = ""

	s.H5.StylePrimitive.Color = stringPtr("#9d7cd8")
	s.H5.StylePrimitive.Prefix = ""

	s.H6.StylePrimitive.Color = stringPtr("#808080")
	s.H6.StylePrimitive.Prefix = ""

	// Inline code: red on dark surface
	s.Code = ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Prefix:          "\u00a0",
			Suffix:          "\u00a0",
			Color:           stringPtr("#e06c75"),
			BackgroundColor: stringPtr("#1e1e1e"),
		},
	}

	// Code blocks: subtle rounded border using box-drawing characters.
	// Top/bottom borders via BlockPrefix/Suffix, left bar via IndentToken.
	// The dim color (#303040) is embedded via ANSI escapes in the strings.
	borderColor := "\033[38;2;48;48;64m" // dim blue-gray
	reset := "\033[0m"
	codeColor := "\033[38;2;238;238;238m" // restore text color after border chars

	topBorder := borderColor + "  ╭───" + reset
	bottomBorder := borderColor + "  ╰───" + reset
	leftBar := borderColor + "│" + reset + codeColor + " "

	s.CodeBlock = ansi.StyleCodeBlock{
		StyleBlock: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color:       stringPtr("#eeeeee"),
				BlockPrefix: "\n" + topBorder + "\n",
				BlockSuffix: bottomBorder + "\n",
			},
			Indent:      uintPtr(1),
			IndentToken: stringPtr(leftBar),
			Margin:      uintPtr(2),
		},
		Chroma: &ansi.Chroma{
			Text:    ansi.StylePrimitive{Color: stringPtr("#eeeeee")},
			Error:   ansi.StylePrimitive{Color: stringPtr("#e06c75")},
			Comment: ansi.StylePrimitive{Color: stringPtr("#606060")},
			CommentPreproc: ansi.StylePrimitive{Color: stringPtr("#f5a742")},
			Keyword:          ansi.StylePrimitive{Color: stringPtr("#9d7cd8")},
			KeywordReserved:  ansi.StylePrimitive{Color: stringPtr("#9d7cd8")},
			KeywordNamespace: ansi.StylePrimitive{Color: stringPtr("#9d7cd8")},
			KeywordType:      ansi.StylePrimitive{Color: stringPtr("#e5c07b")},
			Operator:         ansi.StylePrimitive{Color: stringPtr("#56b6c2")},
			Punctuation:      ansi.StylePrimitive{Color: stringPtr("#eeeeee")},
			Name:             ansi.StylePrimitive{Color: stringPtr("#eeeeee")},
			NameBuiltin:      ansi.StylePrimitive{Color: stringPtr("#f5a742")},
			NameTag:          ansi.StylePrimitive{Color: stringPtr("#e06c75")},
			NameAttribute:    ansi.StylePrimitive{Color: stringPtr("#e5c07b")},
			NameClass:        ansi.StylePrimitive{Color: stringPtr("#e5c07b"), Bold: boolPtr(true)},
			NameDecorator:    ansi.StylePrimitive{Color: stringPtr("#e5c07b")},
			NameFunction:     ansi.StylePrimitive{Color: stringPtr("#5c9cf5")},
			NameConstant:     ansi.StylePrimitive{Color: stringPtr("#f5a742")},
			NameException:    ansi.StylePrimitive{Color: stringPtr("#e06c75")},
			LiteralNumber:    ansi.StylePrimitive{Color: stringPtr("#f5a742")},
			LiteralString:    ansi.StylePrimitive{Color: stringPtr("#7fd88f")},
			LiteralStringEscape: ansi.StylePrimitive{Color: stringPtr("#56b6c2")},
			GenericDeleted:   ansi.StylePrimitive{Color: stringPtr("#e06c75")},
			GenericEmph:      ansi.StylePrimitive{Italic: boolPtr(true)},
			GenericInserted:  ansi.StylePrimitive{Color: stringPtr("#7fd88f")},
			GenericStrong:    ansi.StylePrimitive{Bold: boolPtr(true)},
			GenericSubheading: ansi.StylePrimitive{Color: stringPtr("#5c9cf5")},
		},
	}

	// Links: blue
	s.Link = ansi.StylePrimitive{
		Color:     stringPtr("#56b6c2"),
		Underline: boolPtr(true),
	}
	s.LinkText = ansi.StylePrimitive{
		Color: stringPtr("#56b6c2"),
	}

	// Strong/emphasis
	s.Strong = ansi.StylePrimitive{
		Bold:  boolPtr(true),
		Color: stringPtr("#f5a742"),
	}
	s.Emph = ansi.StylePrimitive{Italic: boolPtr(true)}

	// Horizontal rule: subtle
	s.HorizontalRule = ansi.StylePrimitive{
		Color:  stringPtr("#484848"),
		Format: "\n───\n",
	}

	// Block quotes: muted with left border
	s.BlockQuote = ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Color: stringPtr("#e5c07b"),
		},
		Indent:      uintPtr(1),
		IndentToken: stringPtr("│ "),
	}

	// List items: clean bullet
	s.Item = ansi.StylePrimitive{
		BlockPrefix: "• ",
	}

	return s
}

// renderScrollbar appends a half-block scrollbar to the right edge of content.
// Uses ▐ for track and █/▀/▄ for sub-cell thumb precision, matching OpenCode.
func renderScrollbar(content string, viewportHeight, totalLines, offset int) string {
	if totalLines <= viewportHeight || viewportHeight <= 0 {
		return content
	}

	lines := strings.Split(content, "\n")

	// Calculate in 2x virtual pixel space for sub-cell precision
	virtualHeight := viewportHeight * 2
	virtualThumbSize := int(math.Max(2, math.Round(float64(virtualHeight)*float64(viewportHeight)/float64(totalLines))))
	virtualThumbStart := int(math.Round(float64(offset) * float64(virtualHeight-virtualThumbSize) / float64(totalLines-viewportHeight)))
	virtualThumbEnd := virtualThumbStart + virtualThumbSize

	trackStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#1e1e1e"))
	thumbStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#484848"))

	for i := 0; i < len(lines) && i < viewportHeight; i++ {
		cellTop := i * 2
		cellBottom := cellTop + 1

		topInThumb := cellTop >= virtualThumbStart && cellTop < virtualThumbEnd
		bottomInThumb := cellBottom >= virtualThumbStart && cellBottom < virtualThumbEnd

		var char string
		switch {
		case topInThumb && bottomInThumb:
			char = thumbStyle.Render("█")
		case topInThumb && !bottomInThumb:
			char = thumbStyle.Render("▀")
		case !topInThumb && bottomInThumb:
			char = thumbStyle.Render("▄")
		default:
			char = trackStyle.Render(" ")
		}

		lines[i] = lines[i] + char
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
