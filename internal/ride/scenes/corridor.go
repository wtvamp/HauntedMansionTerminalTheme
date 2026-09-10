package scenes

import (
	"strings"

	"github.com/wtvamp/HauntedMansionTerminalTheme/internal/ride"
)

func init() { ride.Register(corridor{}) }

// corridor is the endless hall of doors. Doors breathe in and out of their
// frames on staggered cycles so the hall never looks like it is on one clock.
type corridor struct{}

func (corridor) Name() string  { return "corridor" }
func (corridor) Title() string { return "The Endless Hallway" }
func (corridor) Ticks() int    { return 56 }

// Each door bulges on its own period. Primes keep them from resynchronising.
var doorPeriods = []int{7, 11, 13, 5, 17}

func (c corridor) Render(f ride.Frame) string {
	const doors = 5
	var rows [7]strings.Builder

	for d := 0; d < doors; d++ {
		bulge := (f.N/2 + d*3) % doorPeriods[d]
		open := bulge < 2

		lintel, body, knob := "┌────┐", "│    │", "│  ○ │"
		if open {
			lintel, body, knob = "┌────┐", "│▒▒▒▒│", "│▒▒○▒│"
		}
		rows[0].WriteString("  " + lintel)
		rows[1].WriteString("  " + body)
		rows[2].WriteString("  " + knob)
		rows[3].WriteString("  " + body)
		rows[4].WriteString("  " + body)
		rows[5].WriteString("  └────┘")
		if open {
			rows[6].WriteString("   knock")
		} else {
			rows[6].WriteString("        ")
		}
	}

	var b strings.Builder
	for i := 0; i < 6; i++ {
		b.WriteString(rows[i].String() + "\n")
	}
	art := f.Style.Paint("ride_velvet", b.String())
	// The knock row is painted separately but must stay in the same column grid
	// as the doors above it, so it is appended as its own line rather than being
	// centred on its own.
	art += "\n" + f.Style.Paint("ride_ectoplasm", strings.TrimRight(rows[6].String(), " "))

	art += "\n\n" + f.Style.Paint("ride_candle", "  Do not stray from the Doom Buggy.")
	return ride.Center(art, f.Width, f.Height)
}
