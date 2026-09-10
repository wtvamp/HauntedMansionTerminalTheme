package scenes

import (
	"math"
	"strings"

	"github.com/wtvamp/HauntedMansionTerminalTheme/internal/ride"
)

func init() { ride.Register(ballroom{}) }

// ballroom waltzes pairs of ghosts along sine paths at 3/4 time.
type ballroom struct{}

func (ballroom) Name() string  { return "ballroom" }
func (ballroom) Title() string { return "The Grand Hall" }
func (ballroom) Ticks() int    { return 72 }

func (b ballroom) Render(f ride.Frame) string {
	const (
		w       = 48
		h       = 9
		dancers = 6
	)
	grid := make([][]rune, h)
	for i := range grid {
		grid[i] = []rune(strings.Repeat(" ", w))
	}

	phase := float64(f.N) * 0.18
	for d := 0; d < dancers; d++ {
		// Each pair orbits a centre offset along the floor, a third of a turn apart.
		off := float64(d) * 2 * math.Pi / dancers
		cx := float64(w) / 2
		cy := float64(h) / 2

		x := cx + math.Cos(phase+off)*(float64(w)/2-6)
		y := cy + math.Sin((phase+off)*1.5)*(float64(h)/2-1.5)

		put(grid, int(math.Round(x)), int(math.Round(y)), '⚭')
		// The partner trails a step behind, which is what makes it read as a waltz
		// rather than as dots on a circle.
		px := cx + math.Cos(phase+off-0.35)*(float64(w)/2-6)
		py := cy + math.Sin((phase+off-0.35)*1.5)*(float64(h)/2-1.5)
		put(grid, int(math.Round(px)), int(math.Round(py)), '◦')
	}

	var sb strings.Builder
	sb.WriteString("╭" + strings.Repeat("─", w) + "╮\n")
	for _, row := range grid {
		sb.WriteString("│" + string(row) + "│\n")
	}
	sb.WriteString("╰" + strings.Repeat("─", w) + "╯")

	art := f.Style.Paint("ride_spectre", sb.String())
	art += "\n\n" + f.Style.Paint("ride_candle", "   Master Gracey's ballroom. The waltz has not stopped since.")
	return ride.Center(art, f.Width, f.Height)
}

func put(grid [][]rune, x, y int, r rune) {
	if y < 0 || y >= len(grid) || x < 0 || x >= len(grid[y]) {
		return
	}
	grid[y][x] = r
}
