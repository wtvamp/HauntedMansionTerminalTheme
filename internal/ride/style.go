package ride

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/wtvamp/HauntedMansionTerminalTheme/internal/palette"
)

// Style resolves palette roles into renderers at a fixed capability.
//
// Scenes ask for a role by name and never touch a hex value or an ANSI number,
// which is what lets the palette stay the single source of truth all the way
// into the animation.
type Style struct {
	Cap      Capability
	palette  *palette.Palette
	renderer *lipgloss.Renderer
	cache    map[string]lipgloss.Style
}

// NewStyle builds the style table for a capability.
func NewStyle(p *palette.Palette, cap Capability, r *lipgloss.Renderer) *Style {
	if r != nil {
		r.SetColorProfile(cap.Profile())
	}
	return &Style{Cap: cap, palette: p, renderer: r, cache: map[string]lipgloss.Style{}}
}

func (s *Style) base() lipgloss.Style {
	if s.renderer != nil {
		return s.renderer.NewStyle()
	}
	return lipgloss.NewStyle()
}

// Role returns a style painting the foreground with the named palette role.
// An unknown role is a programming error and panics, matching palette.MustColor.
func (s *Style) Role(name string) lipgloss.Style {
	if st, ok := s.cache[name]; ok {
		return st
	}
	c := s.palette.MustColor(name)
	st := s.base().Foreground(lipgloss.Color(c.Hex()))
	s.cache[name] = st
	return st
}

// Dim is the style for text that should recede. At 16 colors there is no safe
// dim shade on this background, so it degrades to plain rather than to something
// unreadable.
func (s *Style) Dim() lipgloss.Style {
	if s.Cap <= Ansi16 {
		return s.base()
	}
	return s.Role("greeting_quote")
}

// Paint applies a role to a multi-line block.
//
// It renders line by line on purpose. Handing lipgloss a whole block makes it pad
// every line out to the width of the longest one, which leaves trailing coloured
// whitespace that shows up as ragged blocks against the terminal background and
// breaks Center's alignment maths.
func (s *Style) Paint(role, block string) string {
	st := s.Role(role)
	lines := strings.Split(strings.TrimRight(block, "\n"), "\n")
	for i, l := range lines {
		lines[i] = st.Render(l)
	}
	return strings.Join(lines, "\n")
}
