package ride

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// Frame is everything a scene needs to draw one still image.
//
// Rendering is a pure function of this struct — no scene holds mutable animation
// state. That is what makes `doombuggy --frame graveyard` possible, and what makes
// scenes testable without standing up a terminal.
type Frame struct {
	// N counts ticks since the scene started, from 0.
	N int
	// Progress runs 0.0 to 1.0 across the scene's Ticks.
	Progress float64
	// Width and Height are the usable terminal size.
	Width, Height int
	// Style resolves palette roles.
	Style *Style
	// Seed is fixed for the whole run. Scenes that want variety between runs but
	// stability between frames key off this — Render stays a pure function of
	// Frame, so --frame reproduces exactly what the ride showed.
	Seed int64
}

// Scene is one beat of the ride.
//
// Adding a scene means implementing this and registering it — never editing an
// existing scene.
type Scene interface {
	// Name is the stable identifier used by --frame and the sequence list.
	Name() string
	// Title is what a human sees.
	Title() string
	// Ticks is how many frames the scene runs for before the ride advances.
	Ticks() int
	// Render draws one frame.
	Render(Frame) string
}

// TickInterval is the animation clock. ~12fps: fast enough to read as motion,
// slow enough that a full ride over ssh does not saturate the link.
const TickInterval = 80 * time.Millisecond

var registry = map[string]Scene{}
var order []string

// Register adds a scene. Call it from an init in the scenes package. Registering
// the same name twice is a programming error and panics at startup.
func Register(s Scene) {
	if _, dup := registry[s.Name()]; dup {
		panic(fmt.Sprintf("ride: two scenes named %q", s.Name()))
	}
	registry[s.Name()] = s
	order = append(order, s.Name())
}

// Lookup returns a scene by name.
func Lookup(name string) (Scene, bool) {
	s, ok := registry[name]
	return s, ok
}

// Names lists registered scenes in registration order.
func Names() []string { return append([]string(nil), order...) }

// SortedNames lists them alphabetically, for help output.
func SortedNames() []string {
	out := Names()
	sort.Strings(out)
	return out
}

// Sequence is the ride order. It is explicit rather than "whatever registered
// first" so the order is a decision someone made, visible in one place.
var Sequence = []string{
	"stretching-room",
	"corridor",
	"ballroom",
	"graveyard",
	"hitchhikers",
}

// Scenes resolves Sequence into scenes, skipping the intro if asked.
func Scenes(skipIntro bool) ([]Scene, error) {
	var out []Scene
	for _, name := range Sequence {
		if skipIntro && name == "stretching-room" {
			continue
		}
		s, ok := Lookup(name)
		if !ok {
			return nil, fmt.Errorf("ride sequence names %q, which no scene registered; registered: %s",
				name, strings.Join(SortedNames(), ", "))
		}
		out = append(out, s)
	}
	return out, nil
}

// Center pads a block of text to sit in the middle of w x h.
func Center(block string, w, h int) string {
	lines := strings.Split(block, "\n")
	widest := 0
	for _, l := range lines {
		if n := visibleWidth(l); n > widest {
			widest = n
		}
	}
	left := (w - widest) / 2
	if left < 0 {
		left = 0
	}
	pad := strings.Repeat(" ", left)

	var b strings.Builder
	top := (h - len(lines)) / 2
	for i := 0; i < top; i++ {
		b.WriteByte('\n')
	}
	for i, l := range lines {
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(pad + l)
	}
	return b.String()
}

// visibleWidth counts runes outside SGR escape sequences.
func visibleWidth(s string) int {
	n, inEsc := 0, false
	for _, r := range s {
		switch {
		case r == 0x1b:
			inEsc = true
		case inEsc && r == 'm':
			inEsc = false
		case inEsc:
		default:
			n++
		}
	}
	return n
}
