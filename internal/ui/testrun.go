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

// Styles — uses the OpenCode palette from theme.go.
// Three semantic tiers: pass (green), fail (red), secondary (dim).
var (
	tPass = lipgloss.NewStyle().Foreground(green)
	tFail = lipgloss.NewStyle().Foreground(red)
	tDim  = lipgloss.NewStyle().Foreground(dim)
	tBody = lipgloss.NewStyle().Foreground(white)
	tHead = lipgloss.NewStyle().Foreground(accent).Bold(true)

	tBarPass = lipgloss.NewStyle().Background(green).Foreground(lipgloss.Color("#000000"))
	tBarFail = lipgloss.NewStyle().Background(red).Foreground(lipgloss.Color("#000000"))
	tBarDim  = lipgloss.NewStyle().Background(faint)

	// Highlighted failing line — full-width bg like a pager/editor
	tHlBg = lipgloss.Color("#1a1a2e")
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
	s.viewport.SetContent(s.render())
	return nil
}

// ── Structured output (parsed test cases) ──────────────────────────

func (s *TestRunScreen) render() string {
	r := s.result
	if len(r.Cases) == 0 {
		return s.renderRaw()
	}

	var sections []string

	// 1. Summary box — bordered, contains status + bar + stats
	sections = append(sections, s.renderSummaryBox())

	// 2. Failures — expanded with error detail
	var failures, passes []workspace.TestCase
	for _, tc := range r.Cases {
		if tc.Status == "failed" {
			failures = append(failures, tc)
		} else {
			passes = append(passes, tc)
		}
	}

	if len(failures) > 0 {
		sections = append(sections, s.renderFailures(failures))
	}

	// 3. Passes — tight compact list
	if len(passes) > 0 {
		sections = append(sections, s.renderPasses(passes))
	}

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (s *TestRunScreen) renderSummaryBox() string {
	r := s.result
	var lines []string

	// Status line
	if r.Passed {
		icon := tPass.Bold(true).Render("✓")
		label := tPass.Bold(true).Render("ALL TESTS PASSED")
		lines = append(lines, icon+"  "+label)
	} else {
		icon := tFail.Bold(true).Render("✗")
		label := tFail.Bold(true).Render("TESTS FAILED")
		lines = append(lines, icon+"  "+label)
	}

	// Progress bar
	bar := s.renderBar()
	if bar != "" {
		lines = append(lines, "")
		lines = append(lines, bar)
	}

	// Stats
	var stats []string
	if r.PassCount > 0 {
		stats = append(stats, tPass.Render(fmt.Sprintf("%d passed", r.PassCount)))
	}
	if r.FailCount > 0 {
		stats = append(stats, tFail.Render(fmt.Sprintf("%d failed", r.FailCount)))
	}
	stats = append(stats, tDim.Render(fmt.Sprintf("%d total", r.Total)))
	lines = append(lines, strings.Join(stats, tDim.Render("  ·  ")))

	content := strings.Join(lines, "\n")

	boxWidth := s.width - 6
	if boxWidth < 30 {
		boxWidth = 30
	}

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#333333")).
		Padding(1, 2).
		Width(boxWidth).
		Render(content)

	return "\n" + "  " + box + "\n"
}

func (s *TestRunScreen) renderBar() string {
	r := s.result
	if r.Total == 0 {
		return ""
	}

	barWidth := s.width - 12
	if barWidth < 10 {
		barWidth = 10
	}
	if barWidth > 50 {
		barWidth = 50
	}

	passW := barWidth * r.PassCount / r.Total
	failW := barWidth * r.FailCount / r.Total
	if r.PassCount > 0 && passW == 0 {
		passW = 1
	}
	if r.FailCount > 0 && failW == 0 {
		failW = 1
	}
	emptyW := barWidth - passW - failW
	if emptyW < 0 {
		emptyW = 0
	}

	var bar string
	if passW > 0 {
		bar += tBarPass.Render(strings.Repeat("━", passW))
	}
	if failW > 0 {
		bar += tBarFail.Render(strings.Repeat("━", failW))
	}
	if emptyW > 0 {
		bar += tBarDim.Render(strings.Repeat("━", emptyW))
	}
	return bar
}

func (s *TestRunScreen) renderFailures(failures []workspace.TestCase) string {
	var b strings.Builder

	b.WriteString("  " + tHead.Render("Failures"))
	b.WriteString("\n\n")

	for i, tc := range failures {
		// Icon carries the semantic color; name stays neutral/bright
		icon := tFail.Bold(true).Render("✗")
		name := tBody.Render(tc.Name)
		b.WriteString("    " + icon + "  " + name + "\n")

		if tc.Message != "" {
			for _, line := range strings.Split(tc.Message, "\n") {
				b.WriteString("       " + tDim.Render(line) + "\n")
			}
		}

		if i < len(failures)-1 {
			b.WriteString("\n")
		}
	}

	return b.String()
}

func (s *TestRunScreen) renderPasses(passes []workspace.TestCase) string {
	var b strings.Builder

	b.WriteString("\n  " + tHead.Render("Passed"))
	b.WriteString("\n\n")

	// Tight list — no blank lines between passes
	for _, tc := range passes {
		icon := tPass.Render("✓")
		name := tDim.Render(tc.Name)
		b.WriteString("    " + icon + "  " + name + "\n")
	}

	return b.String()
}

// ── Raw output fallback ────────────────────────────────────────────
// Three color tiers only: fail lines, pass lines, everything else (dim).
// The failing code line gets a background highlight like a pager.

func (s *TestRunScreen) renderRaw() string {
	var b strings.Builder

	// Header with breathing room
	b.WriteString("\n")
	if s.result.Passed {
		icon := tPass.Bold(true).Render("✓")
		label := tPass.Bold(true).Render("ALL TESTS PASSED")
		b.WriteString("    " + icon + "  " + label)
	} else {
		icon := tFail.Bold(true).Render("✗")
		label := tFail.Bold(true).Render("TESTS FAILED")
		b.WriteString("    " + icon + "  " + label)
	}
	b.WriteString("\n")
	sep := tDim.Render(strings.Repeat("─", 40))
	b.WriteString("    " + sep)
	b.WriteString("\n\n")

	hlStyle := lipgloss.NewStyle().Foreground(white).Background(tHlBg).Bold(true)
	caretStyle := lipgloss.NewStyle().Foreground(red).Background(tHlBg)
	descStyle := lipgloss.NewStyle().Foreground(yellow).Bold(true)

	lines := strings.Split(s.result.RawOutput, "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		var styled string

		switch {
		// ── FAIL file line (must be before describe block check) ──
		case strings.HasPrefix(trimmed, "FAIL"):
			styled = "    " + tFail.Render(trimmed)

		// ── Fail-tier: red (× markers) ──
		case isFailLine(trimmed):
			styled = "      " + tFail.Render(trimmed)

		// ── Pass-tier: green (✓ markers) ──
		case isPassLine(trimmed):
			styled = "      " + tPass.Render(trimmed)

		// ── Describe block names (section headers in test tree) ──
		// Identifiers like EXPECTED_MINUTES_IN_OVEN, remainingMinutesInOven
		case isDescribeBlock(trimmed):
			if i > 0 {
				prev := strings.TrimSpace(lines[i-1])
				if prev != "" && !strings.HasPrefix(prev, "FAIL") {
					b.WriteString("\n")
				}
			}
			styled = "    " + descStyle.Render(trimmed)

		// ── Failure detail header (● marker) ──
		case strings.HasPrefix(trimmed, "●"):
			styled = "\n    " + tFail.Bold(true).Render(trimmed)

		// ── Failing code line (> NN |) — pager highlight ──
		case strings.HasPrefix(trimmed, ">") && strings.Contains(trimmed, "|"):
			content := "      " + trimmed
			if s.width > 0 {
				pad := s.width - 6 - len(trimmed)
				if pad > 0 {
					content += strings.Repeat(" ", pad)
				}
			}
			styled = hlStyle.Render(content)

		// ── Caret line (  |  ^) — connect to highlight ──
		case isCaret(trimmed):
			content := "      " + trimmed
			if s.width > 0 {
				pad := s.width - 6 - len(trimmed)
				if pad > 0 {
					content += strings.Repeat(" ", pad)
				}
			}
			styled = caretStyle.Render(content)

		// ── Expected/Received labels ──
		case strings.HasPrefix(trimmed, "Expected:"):
			styled = "      " + tBody.Render(trimmed)
		case strings.HasPrefix(trimmed, "Received:"):
			styled = "      " + tFail.Render(trimmed)

		// ── Code context lines (NN |) — normal weight, not dimmed ──
		case isCodeContext(trimmed):
			styled = "      " + tBody.Render(trimmed)

		// ── Empty lines ──
		case trimmed == "":
			styled = ""

		// ── Everything else — subtle but readable ──
		default:
			styled = "      " + tDim.Render(trimmed)
		}

		b.WriteString(styled + "\n")
	}

	return b.String()
}

// ── Line classification helpers ────────────────────────────────────

func isFailLine(s string) bool {
	if strings.HasPrefix(s, "FAIL") {
		return true
	}
	// Jest uses × (U+00D7), also handle ✕ ✗ ✘
	for _, mark := range []string{"×", "✕", "✗", "✘"} {
		if strings.Contains(s, mark) {
			return true
		}
	}
	return false
}

func isPassLine(s string) bool {
	if strings.HasPrefix(s, "PASS") || strings.HasPrefix(s, "ok ") {
		return true
	}
	for _, mark := range []string{"✓", "✔"} {
		if strings.Contains(s, mark) {
			return true
		}
	}
	return false
}

func isCodeContext(s string) bool {
	if len(s) == 0 {
		return false
	}
	return s[0] >= '0' && s[0] <= '9' && strings.Contains(s, "|")
}

// isDescribeBlock detects Jest describe/context group names in raw output.
// These are identifiers (camelCase, PascalCase, SCREAMING_SNAKE) that don't
// match any other pattern — they're section headers in the test tree.
func isDescribeBlock(s string) bool {
	if len(s) == 0 || len(s) > 80 {
		return false
	}
	// Must be word characters only (letters, digits, underscores)
	for _, r := range s {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') || r == '_') {
			return false
		}
	}
	return true
}

func isCaret(s string) bool {
	if !strings.Contains(s, "^") {
		return false
	}
	cleaned := strings.ReplaceAll(strings.ReplaceAll(s, "|", ""), "^", "")
	return strings.TrimSpace(cleaned) == ""
}

// ── Viewport plumbing ──────────────────────────────────────────────

func (s *TestRunScreen) SetSize(width, height int) {
	s.width = width
	s.height = height
	s.viewport.SetWidth(width)
	s.viewport.SetHeight(height)
	if s.result != nil {
		s.viewport.SetContent(s.render())
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

// ── Feedback screen (submission result) ────────────────────────────

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
			Render("  ✓ " + s.message)
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
