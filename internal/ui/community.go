package ui

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/glamour/v2"
	"charm.land/lipgloss/v2"

	"github.com/rcopra/gym/internal/api"
)

// Messages
type communitySolutionsLoadedMsg struct {
	solutions []api.CommunitySolution
}

type solutionFilesLoadedMsg struct {
	files []api.CommunitySolutionFile
}

type communityErrMsg struct {
	err error
}

// solutionItem adapts api.CommunitySolution for the list.
type solutionItem struct {
	solution api.CommunitySolution
}

func (s solutionItem) Title() string {
	stars := ""
	if s.solution.NumStars > 0 {
		stars = fmt.Sprintf(" ★%d", s.solution.NumStars)
	}
	return fmt.Sprintf("@%s%s", s.solution.Author.Handle, stars)
}

func (s solutionItem) Description() string {
	parts := []string{fmt.Sprintf("%d LOC", s.solution.NumLOC)}
	if s.solution.NumComments > 0 {
		parts = append(parts, fmt.Sprintf("%d comments", s.solution.NumComments))
	}
	return strings.Join(parts, " · ")
}

func (s solutionItem) FilterValue() string { return s.solution.Author.Handle }

// CommunityScreen lists community solutions for an exercise.
type CommunityScreen struct {
	client       *api.Client
	trackSlug    string
	exerciseSlug string
	list         list.Model
	loaded       bool
	err          error
	width        int
	height       int
}

func NewCommunityScreen(client *api.Client, trackSlug, exerciseSlug string) *CommunityScreen {
	delegate := list.NewDefaultDelegate()
	l := list.New(nil, delegate, 0, 0)
	l.Title = "Community Solutions"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.Styles.Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(accent).
		MarginLeft(2)

	l.KeyMap.CursorUp = key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("k", "up"))
	l.KeyMap.CursorDown = key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("j", "down"))
	l.KeyMap.GoToStart = key.NewBinding(key.WithKeys("g"), key.WithHelp("gg", "top"))
	l.KeyMap.GoToEnd = key.NewBinding(key.WithKeys("G"), key.WithHelp("G", "bottom"))
	l.KeyMap.Filter = key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "filter"))

	return &CommunityScreen{
		client:       client,
		trackSlug:    trackSlug,
		exerciseSlug: exerciseSlug,
		list:         l,
	}
}

func (s *CommunityScreen) Init() tea.Cmd {
	if s.loaded {
		return nil
	}
	return s.fetchSolutions
}

func (s *CommunityScreen) fetchSolutions() tea.Msg {
	solutions, err := s.client.GetCommunitySolutions(s.trackSlug, s.exerciseSlug)
	if err != nil {
		return communityErrMsg{err: err}
	}
	return communitySolutionsLoadedMsg{solutions: solutions}
}

func (s *CommunityScreen) SetSize(width, height int) {
	s.width = width
	s.height = height
	s.list.SetSize(width, height)
}

func (s *CommunityScreen) Update(msg tea.Msg) (Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case communitySolutionsLoadedMsg:
		if len(msg.solutions) == 0 {
			s.loaded = true
			s.err = fmt.Errorf("no community solutions available (complete the exercise first to unlock)")
			return s, nil
		}
		items := make([]list.Item, len(msg.solutions))
		for i, sol := range msg.solutions {
			items[i] = solutionItem{solution: sol}
		}
		cmd := s.list.SetItems(items)
		s.loaded = true
		return s, cmd

	case communityErrMsg:
		s.err = msg.err
		s.loaded = true
		return s, nil

	case tea.KeyPressMsg:
		if s.list.FilterState() == list.Filtering {
			break
		}
		switch msg.String() {
		case "q", "esc":
			if s.list.FilterState() == list.FilterApplied {
				s.list.ResetFilter()
				return s, nil
			}
			return s, func() tea.Msg { return PopScreenMsg{} }
		case "enter":
			if s.loaded && s.err == nil {
				if item := s.list.SelectedItem(); item != nil {
					sol := item.(solutionItem)
					viewer := NewSolutionViewerScreen(s.client, s.trackSlug, s.exerciseSlug, sol.solution.Author.Handle)
					return s, pushScreen(viewer)
				}
			}
		}
	}

	var cmd tea.Cmd
	s.list, cmd = s.list.Update(msg)
	return s, cmd
}

func (s *CommunityScreen) View() string {
	if s.err != nil {
		return errStyle.Render(fmt.Sprintf("  %v", s.err))
	}
	if !s.loaded {
		return "  Loading community solutions..."
	}
	return s.list.View()
}

func (s *CommunityScreen) ShortHelp() []key.Binding {
	return []key.Binding{
		key.NewBinding(key.WithKeys("j/k"), key.WithHelp("j/k", "navigate")),
		key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "view")),
	}
}

// SolutionViewerScreen shows the code of a community solution.
type SolutionViewerScreen struct {
	client       *api.Client
	trackSlug    string
	exerciseSlug string
	handle       string
	viewport     viewport.Model
	loaded       bool
	err          error
	width        int
	height       int
}

func NewSolutionViewerScreen(client *api.Client, trackSlug, exerciseSlug, handle string) *SolutionViewerScreen {
	vp := viewport.New()
	return &SolutionViewerScreen{
		client:       client,
		trackSlug:    trackSlug,
		exerciseSlug: exerciseSlug,
		handle:       handle,
		viewport:     vp,
	}
}

func (s *SolutionViewerScreen) Init() tea.Cmd {
	return s.fetchFiles
}

func (s *SolutionViewerScreen) fetchFiles() tea.Msg {
	files, err := s.client.GetCommunitySolutionFiles(s.trackSlug, s.exerciseSlug, s.handle)
	if err != nil {
		return communityErrMsg{err: err}
	}
	return solutionFilesLoadedMsg{files: files}
}

func (s *SolutionViewerScreen) SetSize(width, height int) {
	s.width = width
	s.height = height
	s.viewport.SetWidth(width)
	s.viewport.SetHeight(height)
}

func (s *SolutionViewerScreen) Update(msg tea.Msg) (Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case solutionFilesLoadedMsg:
		s.loaded = true
		s.viewport.SetContent(s.renderFiles(msg.files))
		return s, nil

	case communityErrMsg:
		s.err = msg.err
		s.loaded = true
		return s, nil

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

func (s *SolutionViewerScreen) renderFiles(files []api.CommunitySolutionFile) string {
	if len(files) == 0 {
		return "  No files available"
	}

	var sections []string
	for _, f := range files {
		header := lipgloss.NewStyle().
			Bold(true).
			Foreground(accent).
			Render("  " + f.Filename)

		// Wrap code in markdown fenced block for glamour to syntax highlight
		md := fmt.Sprintf("```\n%s\n```", f.Content)

		width := s.width - 4
		if width < 40 {
			width = 40
		}
		renderer, err := glamour.NewTermRenderer(
			glamourStyleOption(),
			glamour.WithWordWrap(width),
		)
		if err != nil {
			sections = append(sections, header+"\n"+f.Content)
			continue
		}

		rendered, err := renderer.Render(md)
		if err != nil {
			sections = append(sections, header+"\n"+f.Content)
			continue
		}

		sections = append(sections, header+"\n"+rendered)
	}

	author := lipgloss.NewStyle().
		Bold(true).
		Foreground(green).
		Render(fmt.Sprintf("  Solution by @%s", s.handle))

	return author + "\n\n" + strings.Join(sections, "\n")
}

func (s *SolutionViewerScreen) View() string {
	if s.err != nil {
		return errStyle.Render(fmt.Sprintf("  Error: %v", s.err))
	}
	if !s.loaded {
		return "  Loading solution..."
	}
	return s.viewport.View()
}

func (s *SolutionViewerScreen) ShortHelp() []key.Binding {
	return []key.Binding{
		key.NewBinding(key.WithKeys("j/k"), key.WithHelp("j/k", "scroll")),
	}
}
