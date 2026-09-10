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

// The ride draws its own figures rather than reusing the greeting's photographs:
// a photograph cannot slide across the frame or hold a pose, and the scenes are
// animation, not decoration.
const hitchhikerArt = `    .-.        .-.        .-.
   (o o)      (o o)      (o o)
   | O |      | O |      | O |
    \ /        \ /        \ /
   .' '.      .' '.      .' '.
  /     \    /     \    /     \
 '~~~~~~~'  '~~~~~~~'  '~~~~~~~'
    EZRA      PHINEAS      GUS`

func (h hitchhikers) Render(f ride.Frame) string {
	art := f.Style.Paint("ride_ectoplasm", hitchhikerArt)

	// Slide in from the right over the first half, then hold.
	slide := 0
	if f.Progress < 0.5 {
		slide = int((0.5 - f.Progress) * 2 * 40)
	}
	pad := strings.Repeat(" ", slide)

	var b strings.Builder
	for i, line := range strings.Split(art, "\n") {
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(pad + line)
	}
	out := b.String()

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
		out += "\n\n" + f.Style.Paint("ride_velvet", "  Hurry baaaack. Hurry baaaack...")
	}
	return ride.Center(out, f.Width, f.Height)
}
