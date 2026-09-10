// Package scenes holds the ride's scenes. Each file is one self-contained scene
// registered at init; nothing here imports anything else here.
package scenes

import (
	"math"
	"strings"

	"github.com/wtvamp/HauntedMansionTerminalTheme/internal/ride"
)

func init() { ride.Register(stretchingRoom{}) }

// stretchingRoom is the opening beat: a portrait that grows taller as the walls
// scroll upward, then the lights go out.
type stretchingRoom struct{}

func (stretchingRoom) Name() string  { return "stretching-room" }
func (stretchingRoom) Title() string { return "The Stretching Room" }
func (stretchingRoom) Ticks() int    { return 64 }

// portrait is drawn at its shortest. The scene interpolates extra rows into the
// middle of the frame to stretch it, which is why the top and bottom are drawn
// separately from the sides.
var portraitTop = []string{
	"╔══════════════════╗",
	"║   ,-~~~~~~~-,    ║",
	"║  (  o     o  )   ║",
	"║   \\    ‿    /    ║",
	"║    '-.....-'     ║",
}

var portraitBottom = []string{
	"║    /|     |\\     ║",
	"║   / |_____| \\    ║",
	"║  '--'     '--'   ║",
	"╚══════════════════╝",
}

func (s stretchingRoom) Render(f ride.Frame) string {
	// Stretch eases out, so the room lurches at first and then creeps — which is
	// the unsettling half of the effect.
	stretch := int(math.Round(8 * math.Sin(f.Progress*math.Pi/2)))

	var b strings.Builder
	for _, l := range portraitTop {
		b.WriteString(l + "\n")
	}
	for i := 0; i < stretch; i++ {
		b.WriteString("║   :          :   ║\n")
	}
	for i, l := range portraitBottom {
		b.WriteString(l)
		if i < len(portraitBottom)-1 {
			b.WriteByte('\n')
		}
	}

	art := f.Style.Paint("ride_spectre", b.String())

	caption := ""
	switch {
	case f.Progress < 0.3:
		caption = "the room is quite ordinary"
	case f.Progress < 0.6:
		caption = "...or is this room actually stretching?"
	case f.Progress < 0.85:
		caption = "and there are no windows. and no doors."
	default:
		caption = "which offers you this chilling challenge:"
	}
	art += "\n\n" + f.Style.Paint("ride_candle", center(caption, 20))

	return ride.Center(art, f.Width, f.Height)
}

// center pads s to width w so captions of different lengths do not jitter the
// block's overall alignment between frames.
func center(s string, w int) string {
	if len(s) >= w {
		return s
	}
	pad := (w - len(s)) / 2
	return strings.Repeat(" ", pad) + s
}
