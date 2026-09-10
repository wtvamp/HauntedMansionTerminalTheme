// Command doombuggy runs the Haunted Mansion ride in your terminal.
//
//	doombuggy                            # the full ride
//	doombuggy --skip-intro               # straight past the stretching room
//	doombuggy --frame graveyard          # render one scene once and exit
//	doombuggy --frame graveyard --at 40  # ...at a specific tick
//	doombuggy --list                     # what scenes exist
package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"

	"github.com/wtvamp/HauntedMansionTerminalTheme/internal/palette"
	"github.com/wtvamp/HauntedMansionTerminalTheme/internal/ride"
	_ "github.com/wtvamp/HauntedMansionTerminalTheme/internal/ride/scenes"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "doombuggy:", err)
		os.Exit(1)
	}
}

func run() error {
	var (
		skipIntro = flag.Bool("skip-intro", false, "skip the stretching room")
		frame     = flag.String("frame", "", "render one scene and exit, for eyeballing a change")
		at        = flag.Int("at", -1, "with --frame, the tick to render; default is the middle of the scene")
		list      = flag.Bool("list", false, "list scenes and exit")
		seed      = flag.Int64("seed", 0, "fix the run seed; 0 picks one from the clock")
		forceCap  = flag.String("color", "", "override detection: truecolor, 256, 16, none")
	)
	flag.Parse()

	if *list {
		for _, name := range ride.Sequence {
			s, ok := ride.Lookup(name)
			if !ok {
				continue
			}
			fmt.Printf("%-18s %-26s %d ticks\n", s.Name(), s.Title(), s.Ticks())
		}
		return nil
	}

	capability, err := capabilityFrom(*forceCap)
	if err != nil {
		return err
	}
	style := ride.NewStyle(palette.Default(), capability, lipgloss.DefaultRenderer())

	if *seed == 0 {
		*seed = time.Now().UnixNano()
	}

	// --frame writes one still to stdout. It deliberately does not start a tea
	// program: this path has to work in a pipe, in CI, and in a scrollback you
	// can screenshot.
	if *frame != "" {
		s, ok := ride.Lookup(*frame)
		if !ok {
			return fmt.Errorf("no scene named %q; known scenes: %v", *frame, ride.SortedNames())
		}
		n := *at
		if n < 0 {
			n = s.Ticks() / 2
		}
		w, h := terminalSize()
		fmt.Println(s.Render(ride.Frame{
			N:        n,
			Progress: float64(n) / float64(s.Ticks()),
			Width:    w,
			Height:   h - 2,
			Style:    style,
			Seed:     *seed,
		}))
		return nil
	}

	scenes, err := ride.Scenes(*skipIntro)
	if err != nil {
		return err
	}
	if len(scenes) == 0 {
		return fmt.Errorf("no scenes to run")
	}

	rand.Seed(*seed)
	p := tea.NewProgram(ride.NewModel(scenes, style, *seed), tea.WithAltScreen())
	_, err = p.Run()
	return err
}

func capabilityFrom(s string) (ride.Capability, error) {
	switch s {
	case "":
		return ride.DetectEnv(), nil
	case "truecolor":
		return ride.TrueColor, nil
	case "256":
		return ride.Ansi256, nil
	case "16":
		return ride.Ansi16, nil
	case "none":
		return ride.NoColor, nil
	}
	return 0, fmt.Errorf("--color %q: want truecolor, 256, 16 or none", s)
}

// terminalSize falls back to 80x24 when stdout is not a terminal, which is the
// case whenever --frame output is being piped or captured.
func terminalSize() (int, int) {
	w, h, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || w == 0 {
		return 80, 24
	}
	return w, h
}
