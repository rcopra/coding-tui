package ui

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/glamour/v2"
	"charm.land/lipgloss/v2"

	"github.com/rcopra/coding-tui/internal/api"
	"github.com/rcopra/coding-tui/internal/workspace"
)

// Messages
type instructionsLoadedMsg struct {
	content string
}

type hintsLoadedMsg struct {
	content string
}

type downloadedMsg struct {
	dir string
}

type detailErrMsg struct {
	err error
}

// DetailScreen shows exercise instructions rendered as markdown.
type DetailScreen struct {
	client    *api.Client
	workspace *workspace.Workspace
	exercise  api.Exercise
	trackSlug string

	viewport     viewport.Model
	instructions string
	hints        string
	showHints    bool
	downloaded   bool
	downloadDir  string
	loading      bool
	running      bool // test or submit in progress
	err          error
	statusMsg    string
	width        int
	height       int
}

func NewDetailScreen(client *api.Client, ws *workspace.Workspace, exercise api.Exercise, trackSlug string) *DetailScreen {
	vp := viewport.New()
	// Disable d/u half-page bindings — d conflicts with download, u is unexpected
	vp.KeyMap.HalfPageDown.SetEnabled(false)
	vp.KeyMap.HalfPageUp.SetEnabled(false)

	return &DetailScreen{
		client:     client,
		workspace:  ws,
		exercise:   exercise,
		trackSlug:  trackSlug,
		viewport:   vp,
		loading:    true,
		downloaded: ws.IsDownloaded(trackSlug, exercise.Slug),
	}
}

func (s *DetailScreen) Init() tea.Cmd {
	return s.fetchInstructions
}

func (s *DetailScreen) fetchInstructions() tea.Msg {
	content, err := s.workspace.ReadInstructions(s.trackSlug, s.exercise.Slug)
	if err != nil {
		return detailErrMsg{err: err}
	}
	return instructionsLoadedMsg{content: content}
}

func (s *DetailScreen) fetchHints() tea.Msg {
	content, err := s.workspace.ReadHints(s.trackSlug, s.exercise.Slug)
	if err != nil {
		return detailErrMsg{err: err}
	}
	return hintsLoadedMsg{content: content}
}

func (s *DetailScreen) doDownload() tea.Msg {
	dir, err := s.workspace.Download(s.trackSlug, s.exercise.Slug)
	if err != nil {
		return detailErrMsg{err: err}
	}
	return downloadedMsg{dir: dir}
}

func (s *DetailScreen) doRunTests() tea.Msg {
	result, err := s.workspace.RunTests(s.trackSlug, s.exercise.Slug)
	if err != nil {
		return detailErrMsg{err: err}
	}
	return testResultMsg{passed: result.Passed, output: result.Output}
}

func (s *DetailScreen) doSubmit() tea.Msg {
	err := s.workspace.SubmitSolution(s.trackSlug, s.exercise.Slug)
	return submitDoneMsg{err: err}
}

func (s *DetailScreen) openInNvim(newPane bool) tea.Cmd {
	if !s.downloaded {
		s.statusMsg = "Download the exercise first (d)"
		return nil
	}
	filePath, err := s.workspace.SolutionFilePath(s.trackSlug, s.exercise.Slug)
	if err != nil {
		s.statusMsg = fmt.Sprintf("Error: %v", err)
		return nil
	}
	if newPane {
		s.statusMsg = fmt.Sprintf("Opening new pane → %s", filePath)
		return tmuxOpenNewPane(filePath)
	}
	s.statusMsg = fmt.Sprintf("Sent to nvim → %s", filePath)
	return tmuxSendToNvim(filePath)
}

func (s *DetailScreen) SetSize(width, height int) {
	oldWidth := s.width
	s.width = width
	s.height = height
	s.viewport.SetWidth(width - 1) // reserve 1 col for scrollbar
	s.viewport.SetHeight(height - 3) // title header + status line + help bar
	// Re-render markdown if width changed and we have content
	if oldWidth != width && s.instructions != "" {
		s.updateContent()
	}
}

func (s *DetailScreen) renderMarkdown(md string) string {
	// Account for glamour's document margin (2 per side = 4 total) + scrollbar (1)
	glamourGutter := 5
	width := s.width - glamourGutter
	if width < 40 {
		width = 40
	}

	style := exercismGlamourStyle()
	renderer, err := glamour.NewTermRenderer(
		glamour.WithStyles(style),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return md
	}

	rendered, err := renderer.Render(md)
	if err != nil {
		return md
	}
	return rendered
}

// stripFirstH1 removes the first markdown H1 heading (and any blank lines after it)
// since we render the title separately in the header bar.
func stripFirstH1(md string) string {
	lines := strings.Split(md, "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "# ") && !strings.HasPrefix(trimmed, "##") {
			// Remove the H1 line and any immediately following blank lines
			rest := lines[i+1:]
			for len(rest) > 0 && strings.TrimSpace(rest[0]) == "" {
				rest = rest[1:]
			}
			return strings.Join(append(lines[:i], rest...), "\n")
		}
		// Stop looking after we hit non-blank content that isn't an H1
		if trimmed != "" {
			break
		}
	}
	return md
}

func (s *DetailScreen) updateContent() {
	content := stripFirstH1(s.instructions)
	if s.showHints && s.hints != "" {
		content += "\n---\n\n## Hints\n\n" + s.hints
	}
	s.viewport.SetContent(s.renderMarkdown(content))
}

func (s *DetailScreen) Update(msg tea.Msg) (Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case instructionsLoadedMsg:
		s.instructions = msg.content
		s.loading = false
		s.updateContent()
		return s, nil

	case hintsLoadedMsg:
		s.hints = msg.content
		if s.hints == "" {
			s.statusMsg = "No hints available"
		}
		s.updateContent()
		return s, nil

	case downloadedMsg:
		s.downloaded = true
		s.downloadDir = msg.dir
		s.statusMsg = fmt.Sprintf("Downloaded → %s", msg.dir)
		return s, nil

	case tmuxSentMsg:
		if msg.err != nil {
			s.statusMsg = fmt.Sprintf("tmux error: %v (are you in tmux?)", msg.err)
		}
		return s, nil

	case testResultMsg:
		s.running = false
		screen := NewTestRunScreen(s.exercise.Title, msg.passed, msg.output)
		return s, pushScreen(screen)

	case submitDoneMsg:
		s.running = false
		if msg.err != nil {
			s.statusMsg = fmt.Sprintf("Submit failed: %v", msg.err)
			return s, nil
		}
		screen := NewFeedbackScreen(
			fmt.Sprintf("Solution submitted! View at: https://exercism.org/tracks/%s/exercises/%s", s.trackSlug, s.exercise.Slug),
			false,
		)
		return s, pushScreen(screen)

	case detailErrMsg:
		s.err = msg.err
		s.loading = false
		s.running = false
		return s, nil

	case tea.KeyPressMsg:
		if s.running {
			return s, nil // ignore keys while running
		}
		switch msg.String() {
		case "q", "esc":
			return s, func() tea.Msg { return PopScreenMsg{} }
		case "d":
			if !s.downloaded {
				s.statusMsg = "Downloading..."
				return s, s.doDownload
			}
			s.statusMsg = fmt.Sprintf("Already downloaded: %s", s.workspace.ExerciseDir(s.trackSlug, s.exercise.Slug))
			return s, nil
		case "h":
			if s.hints != "" {
				s.showHints = !s.showHints
				s.updateContent()
				return s, nil
			}
			if !s.showHints {
				s.showHints = true
				return s, s.fetchHints
			}
		case "t":
			if !s.downloaded {
				s.statusMsg = "Download the exercise first (d)"
				return s, nil
			}
			s.running = true
			s.statusMsg = "Running tests..."
			return s, s.doRunTests
		case "s":
			if !s.downloaded {
				s.statusMsg = "Download the exercise first (d)"
				return s, nil
			}
			s.running = true
			s.statusMsg = "Submitting..."
			return s, s.doSubmit
		case "e":
			return s, s.openInNvim(false)
		case "E":
			return s, s.openInNvim(true)
		case "c":
			screen := NewCommunityScreen(s.client, s.trackSlug, s.exercise.Slug)
			return s, pushScreen(screen)
		case "o":
			url := fmt.Sprintf("https://exercism.org/tracks/%s/exercises/%s", s.trackSlug, s.exercise.Slug)
			s.statusMsg = fmt.Sprintf("Open in browser: %s", url)
			return s, openBrowser(url)
		}
	}

	var cmd tea.Cmd
	s.viewport, cmd = s.viewport.Update(msg)
	return s, cmd
}

func (s *DetailScreen) View() string {
	if s.err != nil {
		return errStyle.Render(fmt.Sprintf("  Error: %v", s.err))
	}
	if s.loading {
		return "  Loading instructions..."
	}

	// Title header
	title := lipgloss.NewStyle().Bold(true).Foreground(accent).Render(s.exercise.Title)
	header := "  " + title

	// Viewport with scrollbar
	vpContent := s.viewport.View()
	vpContent = renderScrollbar(vpContent, s.viewport.Height(), s.viewport.TotalLineCount(), s.viewport.YOffset())

	status := s.buildStatusLine()
	return header + "\n" + vpContent + "\n" + status
}

func (s *DetailScreen) buildStatusLine() string {
	var parts []string

	if s.downloaded {
		parts = append(parts, special.Render("✓ downloaded"))
	}

	diffStyle, ok := difficultyStyle[s.exercise.Difficulty]
	if !ok {
		diffStyle = subtle
	}
	parts = append(parts, diffStyle.Render(s.exercise.Difficulty))
	parts = append(parts, subtle.Render(s.exercise.Type))

	if s.statusMsg != "" {
		parts = append(parts, lipgloss.NewStyle().Foreground(accent).Render(s.statusMsg))
	}

	line := "  "
	for i, p := range parts {
		if i > 0 {
			line += subtle.Render(" · ")
		}
		line += p
	}
	return line
}

func (s *DetailScreen) ShortHelp() []key.Binding {
	bindings := []key.Binding{
		key.NewBinding(key.WithKeys("j/k"), key.WithHelp("j/k", "scroll")),
	}
	if !s.downloaded {
		bindings = append(bindings, key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "download")))
	} else {
		bindings = append(bindings,
			key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "nvim")),
			key.NewBinding(key.WithKeys("E"), key.WithHelp("E", "nvim (new pane)")),
			key.NewBinding(key.WithKeys("t"), key.WithHelp("t", "test")),
			key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "submit")),
		)
	}
	bindings = append(bindings,
		key.NewBinding(key.WithKeys("h"), key.WithHelp("h", "hints")),
		key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "community")),
		key.NewBinding(key.WithKeys("o"), key.WithHelp("o", "browser")),
	)
	return bindings
}
