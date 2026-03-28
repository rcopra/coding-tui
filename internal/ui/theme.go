package ui

import (
	"charm.land/glamour/v2"
	"charm.land/lipgloss/v2"
)

// Color palette for UI chrome (status bars, badges, etc.)
var (
	white  = lipgloss.Color("#eeeeee") // body text
	dim    = lipgloss.Color("#808080") // muted text
	faint  = lipgloss.Color("#1e1e1e") // surfaces
	accent = lipgloss.Color("#9d7cd8") // purple — headings, primary
	blue   = lipgloss.Color("#5c9cf5") // secondary, links
	green  = lipgloss.Color("#7fd88f") // success, easy
	yellow = lipgloss.Color("#e5c07b") // medium difficulty
	red    = lipgloss.Color("#e06c75") // errors, hard
)

// glamourStyle is the configured style path/name (e.g. "dark", "dracula", "/path/to/style.json").
var glamourStyle = "dark"

// SetGlamourStyle sets the glamour style used for markdown rendering.
func SetGlamourStyle(style string) {
	glamourStyle = style
}

// glamourStyleOption returns the glamour option for the configured style.
func glamourStyleOption() glamour.TermRendererOption {
	return glamour.WithStylePath(glamourStyle)
}
