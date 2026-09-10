package portrait

import (
	"fmt"
	"image"
	"math"
	"strings"

	"github.com/wtvamp/HauntedMansionTerminalTheme/internal/palette"
)

// Color is the palette's colour type. Aliased so callers of this package do not
// have to import both, and so the colour maths stays in one place.
type Color = palette.Color

func deltaE(a, b Color) float64 { return palette.DeltaE(a, b) }

// BlockOptions controls half-block rendering.
type BlockOptions struct {
	// Width in terminal columns.
	Width int
	// Aspect is the cell height-to-width ratio, as for ASCII output.
	Aspect float64
	// Gamma above 1 lifts shadows.
	Gamma float64
	// Quantize maps every pixel to the nearest colour in Palette and emits ANSI
	// slot numbers instead of truecolor. That keeps the art following the
	// terminal's theme and legible at 16 colours, at the cost of exact hues.
	Quantize bool
	// Palette is the 16 ANSI colours, in slot order. Required when Quantize.
	Palette []Color
	// Vignette darkens the edges by this fraction, 0 to disable. Photographs
	// have busy corners that fight the subject at this size.
	Vignette float64
	// Saturation above 1 deepens colour. Dark ride photographs are close to
	// monochrome once brightened, and a little lift puts the mood back.
	Saturation float64
	// AutoContrast stretches the image's own range to full before anything else.
	// Most of these photographs are underexposed interiors.
	AutoContrast bool
}

func BlockDefaults() BlockOptions {
	return BlockOptions{Width: 60, Aspect: 0.5, Gamma: 1.0, Vignette: 0.35, Saturation: 1.0}
}

// upperHalf is drawn with the foreground colour; the background fills the lower
// half. One cell therefore carries two vertically stacked pixels, which is the
// whole reason this looks like a photograph and a character ramp does not.
const upperHalf = "▀"

// Blocks renders an image using half-block glyphs.
func Blocks(img image.Image, o BlockOptions) (string, error) {
	if o.Width < 1 {
		return "", fmt.Errorf("width must be at least 1, got %d", o.Width)
	}
	if o.Aspect <= 0 {
		o.Aspect = 0.5
	}
	if o.Gamma <= 0 {
		o.Gamma = 1
	}
	if o.Quantize && len(o.Palette) == 0 {
		return "", fmt.Errorf("quantising needs a palette")
	}

	b := img.Bounds()
	cols := o.Width
	rows := int(math.Round(float64(cols) * float64(b.Dy()) / float64(b.Dx()) * o.Aspect))
	if rows < 1 {
		rows = 1
	}
	if o.Saturation <= 0 {
		o.Saturation = 1
	}
	// Two samples per cell row.
	sampleRows := rows * 2

	grid := make([]Color, cols*sampleRows)
	for y := 0; y < sampleRows; y++ {
		y0 := b.Min.Y + y*b.Dy()/sampleRows
		y1 := b.Min.Y + (y+1)*b.Dy()/sampleRows
		if y1 <= y0 {
			y1 = y0 + 1
		}
		for x := 0; x < cols; x++ {
			x0 := b.Min.X + x*b.Dx()/cols
			x1 := b.Min.X + (x+1)*b.Dx()/cols
			if x1 <= x0 {
				x1 = x0 + 1
			}
			var rs, gs, bs, n float64
			for yy := y0; yy < y1; yy++ {
				for xx := x0; xx < x1; xx++ {
					cr, cg, cb, _ := img.At(xx, yy).RGBA()
					rs += float64(cr >> 8)
					gs += float64(cg >> 8)
					bs += float64(cb >> 8)
					n++
				}
			}
			c := Color{uint8(rs / n), uint8(gs / n), uint8(bs / n)}
			grid[y*cols+x] = c
		}
	}

	if o.AutoContrast {
		autoContrast(grid)
	}

	for i, c := range grid {
		if o.Gamma != 1 {
			c = adjustGamma(c, o.Gamma)
		}
		if o.Saturation != 1 {
			c = saturate(c, o.Saturation)
		}
		if o.Vignette > 0 {
			c = darken(c, vignetteAt(i%cols, i/cols, cols, sampleRows, o.Vignette))
		}
		grid[i] = c
	}

	var sb strings.Builder
	for y := 0; y < rows; y++ {
		prevFg, prevBg := -1, -1
		var prevFgC, prevBgC Color
		for x := 0; x < cols; x++ {
			top := grid[(y*2)*cols+x]
			bot := grid[(y*2+1)*cols+x]

			if o.Quantize {
				fi, bi := nearestSlot(top, o.Palette), nearestSlot(bot, o.Palette)
				if fi != prevFg || bi != prevBg {
					fmt.Fprintf(&sb, "\x1b[38;5;%d;48;5;%dm", fi, bi)
					prevFg, prevBg = fi, bi
				}
			} else {
				if top != prevFgC || bot != prevBgC || prevFg == -1 {
					fmt.Fprintf(&sb, "\x1b[38;2;%d;%d;%d;48;2;%d;%d;%dm",
						top.R, top.G, top.B, bot.R, bot.G, bot.B)
					prevFgC, prevBgC, prevFg = top, bot, 0
				}
			}
			sb.WriteString(upperHalf)
		}
		sb.WriteString("\x1b[0m\n")
	}
	return sb.String(), nil
}

// autoContrast stretches the grid's luminance range to full. Operating on the
// downsampled grid rather than the source image is deliberate: it is the cells
// that have to be distinguishable, and a bright speck in the original would
// otherwise anchor the white point and flatten everything that matters.
func autoContrast(grid []Color) {
	lo, hi := 1.0, 0.0
	for _, c := range grid {
		l := c.Luminance()
		lo = math.Min(lo, l)
		hi = math.Max(hi, l)
	}
	if hi-lo < 1e-6 {
		return
	}
	scale := 1 / (hi - lo)
	for i, c := range grid {
		f := func(v uint8) uint8 {
			x := (float64(v)/255 - lo) * scale
			return uint8(math.Min(255, math.Max(0, x*255)))
		}
		grid[i] = Color{f(c.R), f(c.G), f(c.B)}
	}
}

// saturate scales chroma around the pixel's own luminance.
func saturate(c Color, f float64) Color {
	l := 0.2126*float64(c.R) + 0.7152*float64(c.G) + 0.0722*float64(c.B)
	g := func(v uint8) uint8 {
		return uint8(math.Min(255, math.Max(0, l+(float64(v)-l)*f)))
	}
	return Color{g(c.R), g(c.G), g(c.B)}
}

// vignetteAt returns how much to darken a cell, 0 at the centre.
func vignetteAt(x, y, w, h int, strength float64) float64 {
	dx := (float64(x)/float64(w-1) - 0.5) * 2
	dy := (float64(y)/float64(h-1) - 0.5) * 2
	d := math.Sqrt(dx*dx+dy*dy) / math.Sqrt2
	return strength * math.Pow(d, 2.2)
}

func darken(c Color, f float64) Color {
	if f <= 0 {
		return c
	}
	if f > 1 {
		f = 1
	}
	k := 1 - f
	return Color{uint8(float64(c.R) * k), uint8(float64(c.G) * k), uint8(float64(c.B) * k)}
}

func adjustGamma(c Color, g float64) Color {
	f := func(v uint8) uint8 {
		return uint8(math.Min(255, math.Pow(float64(v)/255, 1/g)*255))
	}
	return Color{f(c.R), f(c.G), f(c.B)}
}

// nearestSlot finds the closest palette entry, measured in Lab so the match is
// perceptual rather than a naive RGB distance.
func nearestSlot(c Color, pal []Color) int {
	best, bestD := 0, math.MaxFloat64
	for i, p := range pal {
		if d := deltaE(c, p); d < bestD {
			best, bestD = i, d
		}
	}
	return best
}
