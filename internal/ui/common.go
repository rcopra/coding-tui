package ui

import (
	"os/exec"
	"path/filepath"
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

// Styles — high contrast, OpenCode-inspired
var (
	subtle    = lipgloss.NewStyle().Foreground(dim)
	highlight = lipgloss.NewStyle().Foreground(accent)
	special   = lipgloss.NewStyle().Foreground(green)
	errStyle  = lipgloss.NewStyle().Foreground(red)
)

func pushScreen(s Screen) tea.Cmd {
	return func() tea.Msg { return PushScreenMsg{Screen: s} }
}

func popScreen() tea.Msg {
	return PopScreenMsg{}
}

type tmuxSentMsg struct{ err error }

// tmuxOpenNewPane opens a new tmux split to the left with nvim pointing at the file.
func tmuxOpenNewPane(filePath string) tea.Cmd {
	return func() tea.Msg {
		dir := filepath.Dir(filePath)
		// -hb: horizontal split, before (left of current pane)
		cmd := exec.Command("tmux", "split-window", "-hb", "-c", dir, "nvim", filePath)
		err := cmd.Run()
		return tmuxSentMsg{err: err}
	}
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
