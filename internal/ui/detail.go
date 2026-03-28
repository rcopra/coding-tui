package ui

import (
	"errors"
	"fmt"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/glamour/v2"
	"charm.land/lipgloss/v2"

	"github.com/rcopra/gym/internal/api"
	"github.com/rcopra/gym/internal/workspace"
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

type trackNotJoinedMsg struct{}

type completeDoneMsg struct {
	err error
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
	running        bool // test or submit in progress
	confirmSubmit  bool // waiting for second 's' to confirm
	confirmComplete bool // waiting for second 'C' to confirm
	err            error
	statusMsg    string
	width        int
	height       int
}

func NewDetailScreen(client *api.Client, ws *workspace.Workspace, exercise api.Exercise, trackSlug string) *DetailScreen {
	vp := viewport.New()
	// Remap half-page to ctrl+d/ctrl+u (plain d conflicts with download)
	vp.KeyMap.HalfPageDown = key.NewBinding(key.WithKeys("ctrl+d"))
	vp.KeyMap.HalfPageUp = key.NewBinding(key.WithKeys("ctrl+u"))
	// Disable h/l left/right scroll (h conflicts with hints, not useful for markdown)
	vp.KeyMap.Left.SetEnabled(false)
	vp.KeyMap.Right.SetEnabled(false)

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
		if errors.Is(err, api.ErrTrackNotJoined) {
			return trackNotJoinedMsg{}
		}
		return detailErrMsg{err: err}
	}
	return downloadedMsg{dir: dir}
}

func (s *DetailScreen) doRunTests() tea.Msg {
	result, err := s.workspace.RunTests(s.trackSlug, s.exercise.Slug)
	if err != nil {
		return detailErrMsg{err: err}
	}
	return testResultMsg{result: result}
}

func (s *DetailScreen) doSubmit() tea.Msg {
	err := s.workspace.SubmitSolution(s.trackSlug, s.exercise.Slug)
	return submitDoneMsg{err: err}
}

func (s *DetailScreen) doComplete() tea.Msg {
	err := s.workspace.CompleteSolution(s.trackSlug, s.exercise.Slug)
	return completeDoneMsg{err: err}
}

func (s *DetailScreen) openInNvim() tea.Cmd {
	if !s.downloaded {
		s.statusMsg = "Download the exercise first (d)"
		return nil
	}
	filePath, err := s.workspace.SolutionFilePath(s.trackSlug, s.exercise.Slug)
	if err != nil {
		s.statusMsg = fmt.Sprintf("Error: %v", err)
		return nil
	}
	s.statusMsg = fmt.Sprintf("Opening new pane → %s", filePath)
	return tmuxOpenNewPane(filePath)
}

func (s *DetailScreen) SetSize(width, height int) {
	oldWidth := s.width
	s.width = width
	s.height = height
	s.viewport.SetWidth(width)
	s.viewport.SetHeight(height - 3) // title header + status line + help bar
	// Re-render markdown if width changed and we have content
	if oldWidth != width && s.instructions != "" {
		s.updateContent()
	}
}

func (s *DetailScreen) renderMarkdown(md string) string {
	// Account for glamour's document margin (2 per side = 4 total)
	glamourGutter := 4
	width := s.width - glamourGutter
	if width < 40 {
		width = 40
	}

	renderer, err := glamour.NewTermRenderer(
		glamourStyleOption(),
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

	case trackNotJoinedMsg:
		url := fmt.Sprintf("https://exercism.org/tracks/%s", s.trackSlug)
		s.statusMsg = "Track not joined — opening browser to join, then press d again"
		return s, openBrowser(url)

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
		s.statusMsg = ""
		screen := NewTestRunScreen(msg.result, s.workspace, s.trackSlug, s.exercise.Slug)
		return s, pushScreen(screen)

	case submitDoneMsg:
		s.running = false
		s.statusMsg = ""
		if msg.err != nil {
			s.statusMsg = fmt.Sprintf("Submit failed: %v", msg.err)
			return s, nil
		}
		screen := NewFeedbackScreen(
			fmt.Sprintf("Solution submitted! View at: https://exercism.org/tracks/%s/exercises/%s", s.trackSlug, s.exercise.Slug),
			false,
		)
		return s, pushScreen(screen)

	case completeDoneMsg:
		s.running = false
		s.statusMsg = ""
		if msg.err != nil {
			s.statusMsg = fmt.Sprintf("Complete failed: %v", msg.err)
			return s, nil
		}
		screen := NewFeedbackScreen(
			fmt.Sprintf("Exercise completed! https://exercism.org/tracks/%s/exercises/%s", s.trackSlug, s.exercise.Slug),
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
		// Cancel pending confirmation on any key that isn't the confirm key
		k := msg.String()
		if k != "s" && s.confirmSubmit {
			s.confirmSubmit = false
			s.statusMsg = ""
		}
		if k != "C" && s.confirmComplete {
			s.confirmComplete = false
			s.statusMsg = ""
		}
		switch k {
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
			if s.confirmSubmit {
				s.confirmSubmit = false
				s.running = true
				s.statusMsg = "Submitting..."
				return s, s.doSubmit
			}
			s.confirmSubmit = true
			s.confirmComplete = false
			s.statusMsg = "Press s again to submit, any other key to cancel"
			return s, nil
		case "C":
			if !s.downloaded {
				s.statusMsg = "Download the exercise first (d)"
				return s, nil
			}
			if s.confirmComplete {
				s.confirmComplete = false
				s.running = true
				s.statusMsg = "Marking complete..."
				return s, s.doComplete
			}
			s.confirmComplete = true
			s.confirmSubmit = false
			s.statusMsg = "Press C again to mark complete, any other key to cancel"
			return s, nil
		case "e":
			return s, s.openInNvim()
		case "c":
			screen := NewCommunityScreen(s.client, s.trackSlug, s.exercise.Slug)
			return s, pushScreen(screen)
		case "g":
			s.viewport.GotoTop()
			return s, nil
		case "G":
			s.viewport.GotoBottom()
			return s, nil
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

	vpContent := s.viewport.View()

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

	// Scroll position indicator (like Glow)
	total := s.viewport.TotalLineCount()
	vpHeight := s.viewport.Height()
	if total > vpHeight {
		offset := s.viewport.YOffset()
		maxOffset := total - vpHeight
		var pos string
		switch {
		case offset == 0:
			pos = "Top"
		case offset >= maxOffset:
			pos = "Bot"
		default:
			pct := offset * 100 / maxOffset
			pos = fmt.Sprintf("%d%%", pct)
		}
		parts = append(parts, subtle.Render(pos))
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
		key.NewBinding(key.WithKeys("g/G"), key.WithHelp("g/G", "top/bot")),
	}
	if !s.downloaded {
		bindings = append(bindings, key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "download")))
	} else {
		bindings = append(bindings,
			key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "nvim")),
			key.NewBinding(key.WithKeys("t"), key.WithHelp("t", "test")),
			key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "submit")),
			key.NewBinding(key.WithKeys("C"), key.WithHelp("C", "complete")),
		)
	}
	bindings = append(bindings,
		key.NewBinding(key.WithKeys("h"), key.WithHelp("h", "hints")),
		key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "community")),
		key.NewBinding(key.WithKeys("o"), key.WithHelp("o", "browser")),
	)
	return bindings
}
