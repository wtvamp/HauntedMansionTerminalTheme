package scenes

import (
	"strings"

	"github.com/wtvamp/HauntedMansionTerminalTheme/internal/ride"
)

func init() { ride.Register(graveyard{}) }

// graveyard raises tombstones one at a time, each with an epitaph, then lets a
// spectre drift across the front.
type graveyard struct{}

func (graveyard) Name() string  { return "graveyard" }
func (graveyard) Title() string { return "The Graveyard" }
func (graveyard) Ticks() int    { return 80 }

// Epitaphs are original, in the register of the joke tombstones out front: a
// name, a rhyme, and a cause of death that is the punchline.
var epitaphs = [][2]string{
	{"REST IN PEACE", "dear old Fenwick — he tested in prod"},
	{"HERE LIES", "Marion, who force-pushed to main"},
	{"GONE BUT NOT", "forgotten: the branch nobody merged"},
	{"BELOVED", "reviewer — approved without reading"},
	{"IN MEMORIAM", "the on-call who ignored the page"},
}

var stone = []string{
	"   _______   ",
	"  /       \\  ",
	" |  %-5s  | ",
	" |         | ",
	" |   ___   | ",
	"_|_________|_",
}

func (g graveyard) Render(f ride.Frame) string {
	// One stone rises every 12 ticks.
	risen := f.N/12 + 1
	if risen > len(epitaphs) {
		risen = len(epitaphs)
	}

	var rows [6]strings.Builder
	for i := 0; i < risen; i++ {
		for r, line := range stone {
			// A stone rises out of the ground: rows above its current height are
			// blanked, so it appears to push up rather than pop in.
			height := (f.N - i*12) / 2
			if height > 6 {
				height = 6
			}
			if 6-r > height {
				rows[r].WriteString(strings.Repeat(" ", 13))
				continue
			}
			if strings.Contains(line, "%-5s") {
				rows[r].WriteString(strings.Replace(line, "%-5s", "R.I.P", 1))
			} else {
				rows[r].WriteString(line)
			}
		}
	}

	var b strings.Builder
	for i := 0; i < 6; i++ {
		b.WriteString(strings.TrimRight(rows[i].String(), " ") + "\n")
	}
	art := f.Style.Paint("ride_spectre", b.String())

	idx := (risen - 1) % len(epitaphs)
	art += "\n" + f.Style.Paint("ride_candle", "  "+epitaphs[idx][0])
	art += "\n" + f.Style.Paint("ride_ectoplasm", "  "+epitaphs[idx][1])
	return ride.Center(art, f.Width, f.Height)
}
