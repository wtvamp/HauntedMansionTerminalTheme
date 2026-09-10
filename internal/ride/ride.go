package ride

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type tickMsg time.Time

func tick() tea.Cmd {
	return tea.Tick(TickInterval, func(t time.Time) tea.Msg { return tickMsg(t) })
}

// Model drives the sequence of scenes. Scenes themselves are stateless; all the
// state of the ride — which scene, how far into it, how big the window is —
// lives here.
type Model struct {
	scenes  []Scene
	style   *Style
	seed    int64
	current int
	n       int
	w, h    int
	paused  bool
	done    bool
}

// NewModel builds the ride.
func NewModel(scenes []Scene, style *Style, seed int64) Model {
	return Model{scenes: scenes, style: style, seed: seed, w: 80, h: 24}
}

func (m Model) Init() tea.Cmd { return tick() }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.w, m.h = msg.Width, msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			m.done = true
			return m, tea.Quit
		case " ":
			m.paused = !m.paused
			if !m.paused {
				return m, tick()
			}
			return m, nil
		case "right", "n":
			return m.advance()
		case "left", "p":
			if m.current > 0 {
				m.current--
				m.n = 0
			}
			return m, nil
		}

	case tickMsg:
		if m.paused {
			return m, nil
		}
		m.n++
		if m.n >= m.scenes[m.current].Ticks() {
			next, cmd := m.advance()
			if next.done {
				return next, tea.Quit
			}
			return next, cmd
		}
		return m, tick()
	}
	return m, nil
}

func (m Model) advance() (Model, tea.Cmd) {
	m.n = 0
	m.current++
	if m.current >= len(m.scenes) {
		m.current = len(m.scenes) - 1
		m.done = true
		return m, tea.Quit
	}
	return m, tick()
}

func (m Model) View() string {
	if m.done {
		return m.style.Role("ride_ectoplasm").Render("\n  Hurry baaaack. Be sure to bring your death certificate.\n\n")
	}
	s := m.scenes[m.current]

	// Two lines of chrome are reserved at the bottom, so a scene that fills its
	// height does not push the status line off screen.
	body := s.Render(Frame{
		N:        m.n,
		Progress: float64(m.n) / float64(s.Ticks()),
		Width:    m.w,
		Height:   m.h - 2,
		Style:    m.style,
		Seed:     m.seed,
	})
	return body + "\n" + m.status(s)
}

func (m Model) status(s Scene) string {
	left := fmt.Sprintf(" %d/%d  %s", m.current+1, len(m.scenes), s.Title())
	right := "space pause · ←/→ scene · q leave "
	if m.paused {
		left += "  [held]"
	}
	gap := m.w - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
	}
	return m.style.Dim().Render(left + strings.Repeat(" ", gap) + right)
}
