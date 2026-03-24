package ui

import (
	"fmt"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/rcopra/coding-tui/internal/api"
	"github.com/rcopra/coding-tui/internal/workspace"
)

// Messages
type tracksLoadedMsg struct {
	tracks []api.Track
}

type tracksErrMsg struct {
	err error
}

// trackItem adapts api.Track for the bubbles list.
type trackItem struct {
	track api.Track
}

func (t trackItem) Title() string {
	title := t.track.Title

	if t.track.IsJoined {
		title += " " + lipgloss.NewStyle().Foreground(green).Render("✓")
	}
	if t.track.IsNew {
		title += " " + lipgloss.NewStyle().Foreground(accent).Render("new")
	}
	return title
}

func (t trackItem) Description() string {
	desc := fmt.Sprintf("%d exercises · %d concepts", t.track.NumExercises, t.track.NumConcepts)
	if t.track.IsJoined && t.track.NumCompletedExercises > 0 {
		pct := 100 * t.track.NumCompletedExercises / t.track.NumExercises
		desc += fmt.Sprintf(" · %d/%d done (%d%%)", t.track.NumCompletedExercises, t.track.NumExercises, pct)
	}
	return desc
}

func (t trackItem) FilterValue() string { return t.track.Title }

// TracksScreen displays all available tracks.
type TracksScreen struct {
	client    *api.Client
	workspace *workspace.Workspace
	list      list.Model
	loaded    bool
	err       error
	width     int
	height    int
}

func NewTracksScreen(client *api.Client, ws *workspace.Workspace) *TracksScreen {
	delegate := list.NewDefaultDelegate()
	l := list.New(nil, delegate, 0, 0)
	l.Title = "Tracks"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.Styles.Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(accent).
		MarginLeft(2)

	// Vim-style keybindings
	l.KeyMap.CursorUp = key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("k", "up"),
	)
	l.KeyMap.CursorDown = key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("j", "down"),
	)
	l.KeyMap.GoToStart = key.NewBinding(
		key.WithKeys("g"),
		key.WithHelp("gg", "top"),
	)
	l.KeyMap.GoToEnd = key.NewBinding(
		key.WithKeys("G"),
		key.WithHelp("G", "bottom"),
	)
	l.KeyMap.Filter = key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "filter"),
	)

	return &TracksScreen{
		client:    client,
		workspace: ws,
		list:      l,
	}
}

func (s *TracksScreen) Init() tea.Cmd {
	if s.loaded {
		return nil
	}
	return s.fetchTracks
}

func (s *TracksScreen) fetchTracks() tea.Msg {
	tracks, err := s.client.GetTracks()
	if err != nil {
		return tracksErrMsg{err: err}
	}
	return tracksLoadedMsg{tracks: tracks}
}

func (s *TracksScreen) SetSize(width, height int) {
	s.width = width
	s.height = height
	s.list.SetSize(width, height)
}

func (s *TracksScreen) Update(msg tea.Msg) (Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tracksLoadedMsg:
		items := make([]list.Item, len(msg.tracks))
		for i, t := range msg.tracks {
			items[i] = trackItem{track: t}
		}
		cmd := s.list.SetItems(items)
		s.loaded = true
		return s, cmd

	case tracksErrMsg:
		s.err = msg.err
		return s, nil

	case tea.KeyPressMsg:
		// Don't handle navigation keys while filtering
		if s.list.FilterState() == list.Filtering {
			break
		}
		switch msg.String() {
		case "q", "esc":
			if s.list.FilterState() == list.FilterApplied {
				s.list.ResetFilter()
				return s, nil
			}
			// Let root handle quit
			return s, func() tea.Msg { return PopScreenMsg{} }
		case "enter":
			if s.loaded {
				if item := s.list.SelectedItem(); item != nil {
					t := item.(trackItem)
					screen := NewExercisesScreen(s.client, s.workspace, t.track.Slug, t.track.Title)
					return s, pushScreen(screen)
				}
			}
		}
	}

	var cmd tea.Cmd
	s.list, cmd = s.list.Update(msg)
	return s, cmd
}

func (s *TracksScreen) View() string {
	if s.err != nil {
		return errStyle.Render(fmt.Sprintf("  Error loading tracks: %v", s.err))
	}
	if !s.loaded {
		return "  Loading tracks..."
	}
	return s.list.View()
}

func (s *TracksScreen) ShortHelp() []key.Binding {
	return []key.Binding{
		key.NewBinding(key.WithKeys("j/k"), key.WithHelp("j/k", "navigate")),
		key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "filter")),
		key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
	}
}
