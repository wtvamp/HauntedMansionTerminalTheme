package scenes

import (
	"strings"

	contentdata "github.com/wtvamp/HauntedMansionTerminalTheme/content"
	"github.com/wtvamp/HauntedMansionTerminalTheme/internal/ride"
)

func init() { ride.Register(hitchhikers{}) }

// hitchhikers is the closing beat: the three ghosts slide in from the right and
// settle, and the ride signs off with a quote from the shared lore file.
type hitchhikers struct{}

func (hitchhikers) Name() string  { return "hitchhikers" }
func (hitchhikers) Title() string { return "The Hitchhiking Ghosts" }
func (hitchhikers) Ticks() int    { return 64 }

func (h hitchhikers) Render(f ride.Frame) string {
	art, ok := contentdata.Ghost("hitchhikers")
	if !ok {
		// The art is embedded, so this only fires if someone deleted the file —
		// better a plain line than a blank scene.
		art = "three ghosts follow you home"
	}

	// Slide in from the right over the first half, then hold.
	slide := 0
	if f.Progress < 0.5 {
		slide = int((0.5 - f.Progress) * 2 * 40)
	}
	pad := strings.Repeat(" ", slide)

	var b strings.Builder
	for _, line := range strings.Split(art, "\n") {
		b.WriteString(pad + line + "\n")
	}
	out := f.Style.Paint("ride_ectoplasm", strings.TrimRight(b.String(), "\n"))

	if f.Progress > 0.55 {
		quotes := contentdata.Quotes()
		if len(quotes) > 0 {
			// Keyed off the run seed, not the frame counter: a quote that reshuffles
			// every 80ms is unreadable, and one keyed off f.N would never vary
			// because a scene only runs a few dozen ticks.
			q := quotes[int(uint64(f.Seed)%uint64(len(quotes)))]
			out += "\n\n" + f.Style.Paint("ride_candle", "  "+q)
		}
	}
	if f.Progress > 0.8 {
		out += "\n\n" + f.Style.Paint("ride_velvet", "  hurry baaaack...")
	}
	return ride.Center(out, f.Width, f.Height)
}
