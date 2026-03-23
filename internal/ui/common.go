package ui

import (
	"os/exec"
	"runtime"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// Screen is the interface all screens implement.
type Screen interface {
	Init() tea.Cmd
	Update(msg tea.Msg) (Screen, tea.Cmd)
	View() string
	SetSize(width, height int)
	ShortHelp() []key.Binding
}

// Navigation messages sent from screens to the root model.
type PushScreenMsg struct {
	Screen Screen
}

type PopScreenMsg struct{}

// Styles
var (
	subtle    = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	highlight = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
	special   = lipgloss.NewStyle().Foreground(lipgloss.Color("86"))
	errStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
)

func pushScreen(s Screen) tea.Cmd {
	return func() tea.Msg { return PushScreenMsg{Screen: s} }
}

func popScreen() tea.Msg {
	return PopScreenMsg{}
}

type browserOpenedMsg struct{}

func openBrowser(url string) tea.Cmd {
	return func() tea.Msg {
		var cmd *exec.Cmd
		switch runtime.GOOS {
		case "darwin":
			cmd = exec.Command("open", url)
		case "linux":
			cmd = exec.Command("xdg-open", url)
		default:
			cmd = exec.Command("open", url)
		}
		_ = cmd.Start()
		return browserOpenedMsg{}
	}
}
