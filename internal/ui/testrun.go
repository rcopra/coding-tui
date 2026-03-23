package ui

import (
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// Messages
type testResultMsg struct {
	passed bool
	output string
}

type submitDoneMsg struct {
	err error
}

// TestRunScreen displays test output.
type TestRunScreen struct {
	viewport viewport.Model
	passed   bool
	output   string
	title    string
	width    int
	height   int
}

func NewTestRunScreen(title string, passed bool, output string) *TestRunScreen {
	vp := viewport.New()
	s := &TestRunScreen{
		viewport: vp,
		passed:   passed,
		output:   output,
		title:    title,
	}
	return s
}

func (s *TestRunScreen) Init() tea.Cmd {
	s.viewport.SetContent(s.formatOutput())
	return nil
}

func (s *TestRunScreen) formatOutput() string {
	var header string
	if s.passed {
		header = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("86")).
			Render("  ✓ TESTS PASSED")
	} else {
		header = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("196")).
			Render("  ✗ TESTS FAILED")
	}

	// Indent output
	lines := strings.Split(s.output, "\n")
	var indented []string
	for _, line := range lines {
		indented = append(indented, "  "+line)
	}

	return header + "\n\n" + strings.Join(indented, "\n")
}

func (s *TestRunScreen) SetSize(width, height int) {
	s.width = width
	s.height = height
	s.viewport.SetWidth(width)
	s.viewport.SetHeight(height)
}

func (s *TestRunScreen) Update(msg tea.Msg) (Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "esc":
			return s, func() tea.Msg { return PopScreenMsg{} }
		}
	}

	var cmd tea.Cmd
	s.viewport, cmd = s.viewport.Update(msg)
	return s, cmd
}

func (s *TestRunScreen) View() string {
	return s.viewport.View()
}

func (s *TestRunScreen) ShortHelp() []key.Binding {
	return []key.Binding{
		key.NewBinding(key.WithKeys("j/k"), key.WithHelp("j/k", "scroll")),
	}
}

// Feedback screen for showing submission result
type FeedbackScreen struct {
	viewport viewport.Model
	message  string
	isError  bool
	width    int
	height   int
}

func NewFeedbackScreen(message string, isError bool) *FeedbackScreen {
	vp := viewport.New()
	return &FeedbackScreen{
		viewport: vp,
		message:  message,
		isError:  isError,
	}
}

func (s *FeedbackScreen) Init() tea.Cmd {
	var styled string
	if s.isError {
		styled = errStyle.Render("  ✗ " + s.message)
	} else {
		styled = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("86")).
			Render("  ✓ "+s.message)
	}
	s.viewport.SetContent(styled)
	return nil
}

func (s *FeedbackScreen) SetSize(width, height int) {
	s.width = width
	s.height = height
	s.viewport.SetWidth(width)
	s.viewport.SetHeight(height)
}

func (s *FeedbackScreen) Update(msg tea.Msg) (Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "esc", "enter":
			return s, func() tea.Msg { return PopScreenMsg{} }
		}
	}
	return s, nil
}

func (s *FeedbackScreen) View() string {
	return s.viewport.View()
}

func (s *FeedbackScreen) ShortHelp() []key.Binding {
	return []key.Binding{
		key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter/q", "back")),
	}
}
