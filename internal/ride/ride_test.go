package ride_test

import (
	"strings"
	"testing"

	"github.com/wtvamp/HauntedMansionTerminalTheme/internal/palette"
	"github.com/wtvamp/HauntedMansionTerminalTheme/internal/ride"
	_ "github.com/wtvamp/HauntedMansionTerminalTheme/internal/ride/scenes"
)

func TestDetect(t *testing.T) {
	cases := []struct {
		name string
		env  map[string]string
		want ride.Capability
	}{
		{"no-color wins over everything", map[string]string{"NO_COLOR": "1", "COLORTERM": "truecolor", "TERM": "xterm-256color"}, ride.NoColor},
		{"dumb terminal", map[string]string{"TERM": "dumb"}, ride.NoColor},
		{"empty term", map[string]string{}, ride.NoColor},
		{"colorterm truecolor", map[string]string{"TERM": "xterm", "COLORTERM": "truecolor"}, ride.TrueColor},
		{"colorterm 24bit", map[string]string{"TERM": "xterm", "COLORTERM": "24bit"}, ride.TrueColor},
		{"256 in term", map[string]string{"TERM": "xterm-256color"}, ride.Ansi256},
		{"known truecolor emulator", map[string]string{"TERM": "xterm-ghostty"}, ride.TrueColor},
		{"bare xterm falls back to 16", map[string]string{"TERM": "xterm"}, ride.Ansi16},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := ride.Detect(func(k string) string { return c.env[k] })
			if got != c.want {
				t.Errorf("Detect(%v) = %v, want %v", c.env, got, c.want)
			}
		})
	}
}

func TestSequenceIsFullyRegistered(t *testing.T) {
	if _, err := ride.Scenes(false); err != nil {
		t.Fatal(err)
	}
	full, err := ride.Scenes(false)
	if err != nil {
		t.Fatal(err)
	}
	skipped, err := ride.Scenes(true)
	if err != nil {
		t.Fatal(err)
	}
	if len(skipped) != len(full)-1 {
		t.Errorf("--skip-intro dropped %d scenes, want exactly 1", len(full)-len(skipped))
	}
	for _, s := range skipped {
		if s.Name() == "stretching-room" {
			t.Error("--skip-intro still includes the stretching room")
		}
	}
}

// TestScenesRenderEveryTick is the cheap guard that matters most: scenes do
// arithmetic on the frame counter to drive animation, and an off-by-one only
// shows up at the first or last tick. Rendering all of them at every tick and
// every capability costs milliseconds and catches index panics, runaway widths,
// and colors that were only ever tried in truecolor.
func TestScenesRenderEveryTick(t *testing.T) {
	p := palette.Default()
	scenes, err := ride.Scenes(false)
	if err != nil {
		t.Fatal(err)
	}

	const width, height = 100, 30
	for _, capability := range []ride.Capability{ride.TrueColor, ride.Ansi256, ride.Ansi16, ride.NoColor} {
		style := ride.NewStyle(p, capability, nil)
		for _, s := range scenes {
			for n := 0; n <= s.Ticks(); n++ {
				out := s.Render(ride.Frame{
					N:        n,
					Progress: float64(n) / float64(s.Ticks()),
					Width:    width,
					Height:   height,
					Style:    style,
					Seed:     1234,
				})
				if strings.TrimSpace(out) == "" {
					t.Fatalf("%s at %v tick %d rendered nothing", s.Name(), capability, n)
				}
				for i, line := range strings.Split(out, "\n") {
					if w := visible(line); w > width {
						t.Fatalf("%s at %v tick %d line %d is %d columns, terminal is %d",
							s.Name(), capability, n, i, w, width)
					}
				}
			}
		}
	}
}

// TestNoColorEmitsNoEscapes makes the bottom of the degradation ladder real: at
// NoColor there must be no SGR sequences at all, or piping the ride into a file
// produces garbage.
func TestNoColorEmitsNoEscapes(t *testing.T) {
	style := ride.NewStyle(palette.Default(), ride.NoColor, nil)
	scenes, err := ride.Scenes(false)
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range scenes {
		out := s.Render(ride.Frame{N: s.Ticks() / 2, Progress: 0.5, Width: 100, Height: 30, Style: style, Seed: 1})
		if strings.ContainsRune(out, 0x1b) {
			t.Errorf("%s emitted an escape sequence at NoColor", s.Name())
		}
	}
}

func TestCenterHandlesOversizedBlocks(t *testing.T) {
	// A block wider than the terminal must not produce negative padding.
	out := ride.Center(strings.Repeat("x", 200), 40, 10)
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, " ") {
			t.Error("oversized block was still indented")
		}
	}
}

func visible(s string) int {
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
