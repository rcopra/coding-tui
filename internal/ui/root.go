package ui

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type Root struct {
	stack  []Screen
	width  int
	height int
}

func NewRoot(initial Screen) Root {
	return Root{
		stack: []Screen{initial},
	}
}

func (r Root) Init() tea.Cmd {
	if len(r.stack) == 0 {
		return nil
	}
	return r.current().Init()
}

func (r Root) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		r.width = msg.Width
		r.height = msg.Height
		if screen := r.current(); screen != nil {
			screen.SetSize(msg.Width, msg.Height-4) // header + help bar
			updated, cmd := screen.Update(msg)
			r.setCurrent(updated)
			return r, cmd
		}
		return r, nil

	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c":
			return r, tea.Quit
		case "q", "esc":
			// Don't intercept q/esc if the list is filtering
			if s := r.current(); s != nil {
				updated, cmd := s.Update(msg)
				r.setCurrent(updated)
				return r, cmd
			}
			return r, tea.Quit
		}

	case PushScreenMsg:
		msg.Screen.SetSize(r.width, r.height-4)
		r.stack = append(r.stack, msg.Screen)
		return r, r.current().Init()

	case PopScreenMsg:
		if len(r.stack) > 1 {
			r.stack = r.stack[:len(r.stack)-1]
			return r, nil
		}
		return r, tea.Quit
	}

	if screen := r.current(); screen != nil {
		updated, cmd := screen.Update(msg)
		r.setCurrent(updated)
		return r, cmd
	}

	return r, nil
}

func (r Root) View() tea.View {
	if len(r.stack) == 0 {
		return tea.NewView("No screens")
	}

	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(accent).
		Render("  exercism")

	body := r.current().View()
	helpBar := r.renderHelp()

	v := tea.NewView(header + "\n\n" + body + "\n" + helpBar)
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion
	return v
}

func (r Root) renderHelp() string {
	screen := r.current()
	if screen == nil {
		return ""
	}

	bindings := screen.ShortHelp()
	var parts []string
	for _, b := range bindings {
		parts = append(parts, subtle.Render(b.Help().Key)+" "+b.Help().Desc)
	}

	if len(r.stack) > 1 {
		parts = append(parts, subtle.Render("q")+" back")
	} else {
		parts = append(parts, subtle.Render("q")+" quit")
	}

	help := "  "
	for i, p := range parts {
		if i > 0 {
			help += subtle.Render(" · ")
		}
		help += p
	}

	return subtle.Render(help)
}

func (r Root) current() Screen {
	if len(r.stack) == 0 {
		return nil
	}
	return r.stack[len(r.stack)-1]
}

func (r *Root) setCurrent(s Screen) {
	if len(r.stack) > 0 {
		r.stack[len(r.stack)-1] = s
	}
}
