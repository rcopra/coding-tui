package ui

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/rcopra/coding-tui/internal/workspace"
)

// Messages
type testResultMsg struct {
	result *workspace.TestResult
}

type submitDoneMsg struct {
	err error
}

// Styles for test display
var (
	passIcon = lipgloss.NewStyle().Foreground(green).Bold(true).Render("✓")
	failIcon = lipgloss.NewStyle().Foreground(red).Bold(true).Render("✗")
	passText = lipgloss.NewStyle().Foreground(green)
	failText = lipgloss.NewStyle().Foreground(red)
	dimText  = lipgloss.NewStyle().Foreground(dim)
	errText  = lipgloss.NewStyle().Foreground(red).Faint(true)
	barPass  = lipgloss.NewStyle().Background(green).Foreground(lipgloss.Color("#000000"))
	barFail  = lipgloss.NewStyle().Background(red).Foreground(lipgloss.Color("#000000"))
	barDim   = lipgloss.NewStyle().Background(lipgloss.Color("#1e1e1e"))
)

// TestRunScreen displays test output.
type TestRunScreen struct {
	viewport viewport.Model
	result   *workspace.TestResult
	width    int
	height   int
}

func NewTestRunScreen(result *workspace.TestResult) *TestRunScreen {
	vp := viewport.New()
	vp.KeyMap.HalfPageDown.SetEnabled(false)
	vp.KeyMap.HalfPageUp.SetEnabled(false)
	return &TestRunScreen{
		viewport: vp,
		result:   result,
	}
}

func (s *TestRunScreen) Init() tea.Cmd {
	s.viewport.SetContent(s.formatOutput())
	return nil
}

func (s *TestRunScreen) formatOutput() string {
	r := s.result
	var b strings.Builder

	// If we only have raw output, show it nicely
	if len(r.Cases) == 0 {
		return s.formatRawOutput()
	}

	// Summary header
	b.WriteString("\n")
	if r.Passed {
		b.WriteString("  " + passIcon + lipgloss.NewStyle().Bold(true).Foreground(green).Render("  ALL TESTS PASSED"))
	} else {
		b.WriteString("  " + failIcon + lipgloss.NewStyle().Bold(true).Foreground(red).Render("  TESTS FAILED"))
	}
	b.WriteString("\n\n")

	// Progress bar
	b.WriteString(s.renderProgressBar())
	b.WriteString("\n\n")

	// Stats line
	stats := fmt.Sprintf("  %d passed", r.PassCount)
	if r.FailCount > 0 {
		stats += fmt.Sprintf("  %d failed", r.FailCount)
	}
	stats += fmt.Sprintf("  %d total", r.Total)

	statsStyled := "  "
	if r.PassCount > 0 {
		statsStyled += passText.Render(fmt.Sprintf("%d passed", r.PassCount))
	}
	if r.FailCount > 0 {
		if r.PassCount > 0 {
			statsStyled += dimText.Render("  ·  ")
		}
		statsStyled += failText.Render(fmt.Sprintf("%d failed", r.FailCount))
	}
	statsStyled += dimText.Render(fmt.Sprintf("  ·  %d total", r.Total))
	b.WriteString(statsStyled)
	b.WriteString("\n")

	// Separator
	b.WriteString("\n")

	// Individual test cases — failures first
	var failures, passes []workspace.TestCase
	for _, tc := range r.Cases {
		if tc.Status == "failed" {
			failures = append(failures, tc)
		} else {
			passes = append(passes, tc)
		}
	}

	// Show failures
	if len(failures) > 0 {
		b.WriteString(failText.Bold(true).Render("  Failures"))
		b.WriteString("\n\n")
		for _, tc := range failures {
			b.WriteString("  " + failIcon + "  " + lipgloss.NewStyle().Foreground(white).Render(tc.Name))
			b.WriteString("\n")
			if tc.Message != "" {
				for _, line := range strings.Split(tc.Message, "\n") {
					b.WriteString("     " + errText.Render(line))
					b.WriteString("\n")
				}
			}
			b.WriteString("\n")
		}
	}

	// Show passes
	if len(passes) > 0 {
		if len(failures) > 0 {
			b.WriteString("\n")
		}
		b.WriteString(passText.Bold(true).Render("  Passed"))
		b.WriteString("\n\n")
		for _, tc := range passes {
			b.WriteString("  " + passIcon + "  " + dimText.Render(tc.Name))
			b.WriteString("\n")
		}
	}

	return b.String()
}

func (s *TestRunScreen) renderProgressBar() string {
	r := s.result
	if r.Total == 0 {
		return ""
	}

	barWidth := s.width - 6
	if barWidth < 10 {
		barWidth = 10
	}
	if barWidth > 60 {
		barWidth = 60
	}

	passWidth := barWidth * r.PassCount / r.Total
	failWidth := barWidth * r.FailCount / r.Total
	// Ensure at least 1 char for non-zero counts
	if r.PassCount > 0 && passWidth == 0 {
		passWidth = 1
	}
	if r.FailCount > 0 && failWidth == 0 {
		failWidth = 1
	}
	emptyWidth := barWidth - passWidth - failWidth
	if emptyWidth < 0 {
		emptyWidth = 0
	}

	bar := "  "
	if passWidth > 0 {
		bar += barPass.Render(strings.Repeat("━", passWidth))
	}
	if failWidth > 0 {
		bar += barFail.Render(strings.Repeat("━", failWidth))
	}
	if emptyWidth > 0 {
		bar += barDim.Render(strings.Repeat("━", emptyWidth))
	}

	return bar
}

func (s *TestRunScreen) formatRawOutput() string {
	var b strings.Builder

	b.WriteString("\n")
	if s.result.Passed {
		b.WriteString("  " + passIcon + lipgloss.NewStyle().Bold(true).Foreground(green).Render("  TESTS PASSED"))
	} else {
		b.WriteString("  " + failIcon + lipgloss.NewStyle().Bold(true).Foreground(red).Render("  TESTS FAILED"))
	}
	b.WriteString("\n\n")

	// Colorize raw output lines
	for _, line := range strings.Split(s.result.RawOutput, "\n") {
		trimmed := strings.TrimSpace(line)
		styled := "  " + line

		switch {
		// Failed test lines: × ✕ ✗ ✘ markers (Jest uses ×)
		case strings.Contains(trimmed, "×") || strings.Contains(trimmed, "✕") ||
			strings.Contains(trimmed, "✗") || strings.Contains(trimmed, "✘"):
			styled = "  " + failText.Render(line)
		// Passed test lines: ✓ ✔ markers
		case strings.Contains(trimmed, "✓") || strings.Contains(trimmed, "✔"):
			styled = "  " + passText.Render(line)
		// Failure detail headers: ● marker
		case strings.HasPrefix(trimmed, "●"):
			styled = "  " + lipgloss.NewStyle().Foreground(red).Bold(true).Render(line)
		// PASS/ok lines
		case strings.HasPrefix(trimmed, "PASS") || strings.HasPrefix(trimmed, "ok ") ||
			strings.Contains(trimmed, "0 failures"):
			styled = "  " + passText.Render(line)
		// FAIL lines
		case strings.HasPrefix(trimmed, "FAIL"):
			styled = "  " + failText.Render(line)
		// Expected values
		case strings.HasPrefix(trimmed, "Expected:"):
			styled = "  " + passText.Render(line)
		// Received values
		case strings.HasPrefix(trimmed, "Received:"):
			styled = "  " + failText.Render(line)
		// Code pointer lines (> 11 |  code)
		case strings.HasPrefix(trimmed, ">") && strings.Contains(trimmed, "|"):
			styled = "  " + lipgloss.NewStyle().Foreground(yellow).Render(line)
		// Line number context (  12 |  })
		case len(trimmed) > 0 && trimmed[0] >= '0' && trimmed[0] <= '9' && strings.Contains(trimmed, "|"):
			styled = "  " + dimText.Render(line)
		// Caret pointer line (  |  ^)
		case strings.Contains(trimmed, "|") && strings.Contains(trimmed, "^") &&
			strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(trimmed, "|", ""), "^", "")) == "":
			styled = "  " + failText.Render(line)
		// Stack trace / location lines
		case strings.HasPrefix(trimmed, "at "):
			styled = "  " + dimText.Render(line)
		// Assertion descriptions (expect().toBe(), etc.)
		case strings.Contains(trimmed, "expect(") && strings.Contains(trimmed, "//"):
			styled = "  " + dimText.Render(line)
		// General error/fail keywords
		case strings.Contains(trimmed, "Error") || strings.Contains(trimmed, "error") ||
			strings.Contains(trimmed, "failed"):
			styled = "  " + failText.Render(line)
		case strings.Contains(trimmed, "passed"):
			styled = "  " + passText.Render(line)
		// Comments and separators
		case strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "---") ||
			strings.HasPrefix(trimmed, "//"):
			styled = "  " + dimText.Render(line)
		}

		b.WriteString(styled + "\n")
	}

	return b.String()
}

func (s *TestRunScreen) SetSize(width, height int) {
	s.width = width
	s.height = height
	s.viewport.SetWidth(width)
	s.viewport.SetHeight(height)
	// Re-render if we have content (progress bar depends on width)
	if s.result != nil {
		s.viewport.SetContent(s.formatOutput())
	}
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
			Foreground(green).
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
