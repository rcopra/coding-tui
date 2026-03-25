package ui

import (
	"fmt"
	"sort"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/rcopra/coding-tui/internal/api"
	"github.com/rcopra/coding-tui/internal/workspace"
)

// Messages
type exercisesLoadedMsg struct {
	exercises []api.Exercise
}

type exercisesErrMsg struct {
	err error
}

// Difficulty colors
var difficultyStyle = map[string]lipgloss.Style{
	"easy":   lipgloss.NewStyle().Foreground(green),
	"medium": lipgloss.NewStyle().Foreground(yellow),
	"hard":   lipgloss.NewStyle().Foreground(red),
}

// Exercise type indicators
var typeIndicator = map[string]string{
	"tutorial": "★",
	"concept":  "◆",
	"practice": "●",
}

// Sort modes for the exercise list.
type sortMode int

const (
	sortDefault sortMode = iota
	sortDifficulty
	sortAlphabetical
	sortType
	sortModeCount // sentinel for cycling
)

var sortModeLabel = map[sortMode]string{
	sortDefault:      "default",
	sortDifficulty:   "difficulty",
	sortAlphabetical: "a-z",
	sortType:         "type",
}

var difficultyRank = map[string]int{
	"easy":   0,
	"medium": 1,
	"hard":   2,
}

var typeRank = map[string]int{
	"tutorial": 0,
	"concept":  1,
	"practice": 2,
}

// exerciseItem adapts api.Exercise for the bubbles list.
type exerciseItem struct {
	exercise api.Exercise
}

func (e exerciseItem) Title() string {
	indicator := typeIndicator[e.exercise.Type]
	if indicator == "" {
		indicator = "●"
	}

	diff := e.exercise.Difficulty
	style, ok := difficultyStyle[diff]
	if !ok {
		style = subtle
	}

	title := fmt.Sprintf("%s %s", indicator, e.exercise.Title)

	if !e.exercise.IsUnlocked {
		title = "🔒 " + e.exercise.Title
	}

	return title + "  " + style.Render(diff)
}

func (e exerciseItem) Description() string {
	return e.exercise.Blurb
}

func (e exerciseItem) FilterValue() string { return e.exercise.Title }

// ExercisesScreen displays exercises for a given track.
type ExercisesScreen struct {
	client    *api.Client
	workspace *workspace.Workspace
	trackSlug string
	trackName string
	list      list.Model
	exercises []api.Exercise
	sortMode  sortMode
	loaded    bool
	err       error
	width     int
	height    int
}

func NewExercisesScreen(client *api.Client, ws *workspace.Workspace, trackSlug, trackName string) *ExercisesScreen {
	delegate := list.NewDefaultDelegate()
	l := list.New(nil, delegate, 0, 0)
	l.Title = trackName
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

	return &ExercisesScreen{
		client:    client,
		workspace: ws,
		trackSlug: trackSlug,
		trackName: trackName,
		list:      l,
	}
}

func (s *ExercisesScreen) Init() tea.Cmd {
	if s.loaded {
		return nil
	}
	return s.fetchExercises
}

func (s *ExercisesScreen) fetchExercises() tea.Msg {
	exercises, err := s.client.GetExercises(s.trackSlug)
	if err != nil {
		return exercisesErrMsg{err: err}
	}
	return exercisesLoadedMsg{exercises: exercises}
}

func (s *ExercisesScreen) SetSize(width, height int) {
	s.width = width
	s.height = height
	s.list.SetSize(width, height)
}

func (s *ExercisesScreen) Update(msg tea.Msg) (Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case exercisesLoadedMsg:
		s.exercises = msg.exercises
		s.loaded = true
		cmd := s.applySort()
		return s, cmd

	case exercisesErrMsg:
		s.err = msg.err
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
		case "S":
			if s.loaded {
				s.sortMode = (s.sortMode + 1) % sortModeCount
				s.updateTitle()
				cmd := s.applySort()
				return s, cmd
			}
		case "enter":
			if s.loaded {
				if item := s.list.SelectedItem(); item != nil {
					e := item.(exerciseItem)
					screen := NewDetailScreen(s.client, s.workspace, e.exercise, s.trackSlug)
					return s, pushScreen(screen)
				}
			}
		}
	}

	var cmd tea.Cmd
	s.list, cmd = s.list.Update(msg)
	return s, cmd
}

func (s *ExercisesScreen) View() string {
	if s.err != nil {
		return errStyle.Render(fmt.Sprintf("  Error loading exercises: %v", s.err))
	}
	if !s.loaded {
		return "  Loading exercises..."
	}
	return s.list.View()
}

// exerciseSortKey returns a numeric priority for sorting exercises:
// 0 = recommended + in progress, 1 = recommended, 2 = in progress,
// 3 = unlocked, 4 = locked.
func (s *ExercisesScreen) exerciseSortKey(e api.Exercise) int {
	inProgress := s.workspace.IsDownloaded(s.trackSlug, e.Slug)
	switch {
	case e.IsRecommended && inProgress:
		return 0
	case inProgress:
		return 1
	case e.IsRecommended:
		return 2
	case e.IsUnlocked:
		return 3
	default:
		return 4
	}
}

func (s *ExercisesScreen) applySort() tea.Cmd {
	sorted := make([]api.Exercise, len(s.exercises))
	copy(sorted, s.exercises)

	switch s.sortMode {
	case sortDifficulty:
		sort.SliceStable(sorted, func(i, j int) bool {
			return difficultyRank[sorted[i].Difficulty] < difficultyRank[sorted[j].Difficulty]
		})
	case sortAlphabetical:
		sort.SliceStable(sorted, func(i, j int) bool {
			return sorted[i].Title < sorted[j].Title
		})
	case sortType:
		sort.SliceStable(sorted, func(i, j int) bool {
			return typeRank[sorted[i].Type] < typeRank[sorted[j].Type]
		})
	default:
		sort.SliceStable(sorted, func(i, j int) bool {
			return s.exerciseSortKey(sorted[i]) < s.exerciseSortKey(sorted[j])
		})
	}

	items := make([]list.Item, len(sorted))
	for i, e := range sorted {
		items[i] = exerciseItem{exercise: e}
	}
	return s.list.SetItems(items)
}

func (s *ExercisesScreen) updateTitle() {
	if s.sortMode == sortDefault {
		s.list.Title = s.trackName
	} else {
		s.list.Title = fmt.Sprintf("%s · %s", s.trackName, sortModeLabel[s.sortMode])
	}
}

func (s *ExercisesScreen) ShortHelp() []key.Binding {
	return []key.Binding{
		key.NewBinding(key.WithKeys("j/k"), key.WithHelp("j/k", "navigate")),
		key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "filter")),
		key.NewBinding(key.WithKeys("S"), key.WithHelp("S", "sort")),
		key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
	}
}
