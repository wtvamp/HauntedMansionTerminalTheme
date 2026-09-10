// Package portrait converts images into ASCII art sized for a terminal.
//
// The output is meant to sit in content/ghosts/, so it is plain ASCII, has no
// escape sequences, and is measured in terminal cells rather than pixels.
package portrait

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"strings"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
)

// Ramps map brightness to characters, darkest first. On a dark terminal a dense
// glyph reads as bright, so these run space -> @ rather than the other way.
var Ramps = map[string]string{
	"dense":  " .'`^\",:;Il!i><~+_-?][}{1)(|\\/tfjrxnuvczXYUJCLQ0OZmwqpdbkhao*#MW&8%B@$",
	"medium": " .:-=+*#%@",
	"light":  " .:*oO@",
	"blocks": " ░▒▓█",
}

// RampNames lists ramps in a stable order for help text.
var RampNames = []string{"dense", "medium", "light", "blocks"}

// Options controls a conversion.
type Options struct {
	// Width in terminal columns.
	Width int
	// Ramp is a key in Ramps.
	Ramp string
	// RampChars is a literal ramp, darkest first. When set it wins over Ramp.
	//
	// Kept separate from Ramp on purpose: letting one field mean "a name, or
	// else the characters themselves" turns a typo'd name into a silently valid
	// ramp made of its own letters, and the output is garbage with no error.
	RampChars string
	// Aspect is the cell height-to-width ratio. Terminal cells are roughly twice
	// as tall as they are wide, so 0.5 keeps a picture from stretching vertically.
	Aspect float64
	// Gamma above 1 lifts shadows, below 1 deepens them.
	Gamma float64
	// Normalize stretches the image's own brightness range to full black-to-white
	// before mapping. Without it, a low-contrast source lands in a few ramp steps
	// and comes out as mush.
	Normalize bool
	// Black and White clip the input range before mapping: anything at or below
	// Black becomes the first ramp character, anything at or above White becomes
	// the last. Without a black point a rendered vignette's near-black surround
	// lands one step up the ramp and speckles the whole background with dots.
	Black, White float64
	// Invert swaps light and dark, for art meant to sit on a light background.
	Invert bool
	// Trim removes uniformly blank rows and columns from the edges.
	Trim bool
}

// Defaults are tuned for a dark terminal.
func Defaults() Options {
	return Options{Width: 60, Ramp: "medium", Aspect: 0.5, Gamma: 1.0,
		Normalize: true, Black: 0.06, White: 1.0, Trim: true}
}

func (o Options) ramp() (string, error) {
	if o.RampChars != "" {
		if len(o.RampChars) < 2 {
			return "", fmt.Errorf("a literal ramp needs at least two characters, got %q", o.RampChars)
		}
		return o.RampChars, nil
	}
	if r, ok := Ramps[o.Ramp]; ok {
		return r, nil
	}
	return "", fmt.Errorf("unknown ramp %q; known ramps are %s",
		o.Ramp, strings.Join(RampNames, ", "))
}

// luminance is perceptual brightness in 0..1 from a color.
func luminance(c color.Color) float64 {
	r, g, b, a := c.RGBA()
	if a == 0 {
		return 0 // treat transparent as background
	}
	// Un-premultiply so a semi-transparent pixel is not darkened twice.
	fr := float64(r) / float64(a)
	fg := float64(g) / float64(a)
	fb := float64(b) / float64(a)
	l := 0.2126*fr + 0.7152*fg + 0.0722*fb
	return math.Min(1, math.Max(0, l))
}

// Convert renders an image as ASCII.
func Convert(img image.Image, o Options) (string, error) {
	ramp, err := o.ramp()
	if err != nil {
		return "", err
	}
	if o.Width < 1 {
		return "", fmt.Errorf("width must be at least 1, got %d", o.Width)
	}
	if o.Aspect <= 0 {
		o.Aspect = 0.5
	}
	if o.Gamma <= 0 {
		o.Gamma = 1
	}

	b := img.Bounds()
	if b.Dx() == 0 || b.Dy() == 0 {
		return "", fmt.Errorf("image is empty")
	}
	cols := o.Width
	rows := int(math.Round(float64(cols) * float64(b.Dy()) / float64(b.Dx()) * o.Aspect))
	if rows < 1 {
		rows = 1
	}

	// Box-average each cell rather than point-sampling: a thin bright line in the
	// source should darken its whole cell, not vanish or blow it out.
	cells := make([]float64, cols*rows)
	for y := 0; y < rows; y++ {
		y0 := b.Min.Y + y*b.Dy()/rows
		y1 := b.Min.Y + (y+1)*b.Dy()/rows
		if y1 <= y0 {
			y1 = y0 + 1
		}
		for x := 0; x < cols; x++ {
			x0 := b.Min.X + x*b.Dx()/cols
			x1 := b.Min.X + (x+1)*b.Dx()/cols
			if x1 <= x0 {
				x1 = x0 + 1
			}
			var sum float64
			var n int
			for yy := y0; yy < y1; yy++ {
				for xx := x0; xx < x1; xx++ {
					sum += luminance(img.At(xx, yy))
					n++
				}
			}
			cells[y*cols+x] = sum / float64(n)
		}
	}

	if o.Normalize {
		lo, hi := 1.0, 0.0
		for _, v := range cells {
			lo = math.Min(lo, v)
			hi = math.Max(hi, v)
		}
		if hi-lo > 1e-6 {
			for i, v := range cells {
				cells[i] = (v - lo) / (hi - lo)
			}
		}
	}

	if o.White <= o.Black {
		return "", fmt.Errorf("white point %.3f must be above black point %.3f", o.White, o.Black)
	}

	var sb strings.Builder
	last := float64(len(ramp) - 1)
	for y := 0; y < rows; y++ {
		for x := 0; x < cols; x++ {
			v := (cells[y*cols+x] - o.Black) / (o.White - o.Black)
			v = math.Min(1, math.Max(0, v))
			if o.Gamma != 1 {
				v = math.Pow(v, 1/o.Gamma)
			}
			if o.Invert {
				v = 1 - v
			}
			i := int(math.Round(v * last))
			if i < 0 {
				i = 0
			}
			if i > len(ramp)-1 {
				i = len(ramp) - 1
			}
			sb.WriteByte(ramp[i])
		}
		sb.WriteByte('\n')
	}

	out := sb.String()
	if o.Trim {
		out = trim(out)
	}
	return out, nil
}

// trim drops blank edge rows and the common leading/trailing blank columns, so
// art lands flush and Center has an honest width to work with.
func trim(s string) string {
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	blank := func(l string) bool { return strings.TrimSpace(l) == "" }
	for len(lines) > 0 && blank(lines[0]) {
		lines = lines[1:]
	}
	for len(lines) > 0 && blank(lines[len(lines)-1]) {
		lines = lines[:len(lines)-1]
	}
	if len(lines) == 0 {
		return ""
	}
	left := math.MaxInt32
	for _, l := range lines {
		if blank(l) {
			continue
		}
		n := len(l) - len(strings.TrimLeft(l, " "))
		if n < left {
			left = n
		}
	}
	if left == math.MaxInt32 {
		left = 0
	}
	for i, l := range lines {
		if len(l) >= left {
			l = l[left:]
		}
		lines[i] = strings.TrimRight(l, " ")
	}
	return strings.Join(lines, "\n") + "\n"
}

// Region is a tagged area of a rendered portrait. The renderer stamps these
// into a companion image; the key is what gets stored beside the art.
//
// Keep in step with the constants at the top of tools/render-portraits/lib.py.
type Region struct {
	Value byte   // the grey level the renderer stamps
	Key   byte   // the character stored in the .map file
	Role  string // the palette role it resolves to
}

// REGIONS is the whole vocabulary. A cell that matches nothing lands on ground.
var REGIONS = []Region{
	{0, '.', "art_ground"},
	{40, 'f', "art_frame"},
	{80, 'c', "art_cloth"},
	{120, 'h', "art_hair"},
	{160, 's', "art_skin"},
	{200, 'g', "art_glow"},
	{240, 'a', "art_accent"},
}

// RoleForKey resolves a map character to a palette role.
func RoleForKey(k byte) (string, bool) {
	for _, r := range REGIONS {
		if r.Key == k {
			return r.Role, true
		}
	}
	return "", false
}

// nearestRegion picks the region whose stamp value is closest to v. Downsampling
// averages neighbouring stamps, so an exact match is not guaranteed and the
// nearest one is the honest answer.
func nearestRegion(v float64) Region {
	best, bestD := REGIONS[0], math.MaxFloat64
	for _, r := range REGIONS {
		if d := math.Abs(v*255 - float64(r.Value)); d < bestD {
			best, bestD = r, d
		}
	}
	return best
}

// ConvertWithRegions renders the art and, from a companion region image, a map
// of the same shape naming each cell's palette role.
//
// The two images must have the same aspect; the region image is sampled with a
// mode filter rather than an average, because averaging region ids produces ids
// that mean something else entirely.
func ConvertWithRegions(img, regions image.Image, o Options) (art string, colorMap string, err error) {
	art, err = Convert(img, o)
	if err != nil {
		return "", "", err
	}
	lines := strings.Split(strings.TrimRight(art, "\n"), "\n")
	if len(lines) == 0 {
		return art, "", nil
	}

	// Convert re-trims, so map the region image on the same grid as the art by
	// running it through an untrimmed pass and cropping identically.
	untrimmed := o
	untrimmed.Trim = false
	full, err := Convert(img, untrimmed)
	if err != nil {
		return "", "", err
	}
	fullLines := strings.Split(strings.TrimRight(full, "\n"), "\n")
	rowOff, colOff := offsets(fullLines, lines)

	b := regions.Bounds()
	cols, rows := len(fullLines[0]), len(fullLines)
	var mb strings.Builder
	for y := 0; y < len(lines); y++ {
		for x := 0; x < len(lines[y]); x++ {
			if lines[y][x] == ' ' {
				mb.WriteByte('.')
				continue
			}
			gy, gx := y+rowOff, x+colOff
			// Mode over the cell: the most common region wins, so a thin accent
			// does not get averaged into the region next to it.
			counts := map[byte]int{}
			y0 := b.Min.Y + gy*b.Dy()/rows
			y1 := b.Min.Y + (gy+1)*b.Dy()/rows
			x0 := b.Min.X + gx*b.Dx()/cols
			x1 := b.Min.X + (gx+1)*b.Dx()/cols
			for yy := y0; yy < max(y1, y0+1); yy++ {
				for xx := x0; xx < max(x1, x0+1); xx++ {
					counts[nearestRegion(luminance(regions.At(xx, yy))).Key]++
				}
			}
			bestKey, bestN := byte('.'), -1
			for _, r := range REGIONS { // stable order, so ties resolve the same way every run
				if n := counts[r.Key]; n > bestN {
					bestKey, bestN = r.Key, n
				}
			}
			mb.WriteByte(bestKey)
		}
		mb.WriteByte('\n')
	}
	return art, mb.String(), nil
}

// offsets finds where the trimmed art sits inside the untrimmed grid.
func offsets(full, trimmed []string) (row, col int) {
	for i, l := range full {
		if strings.TrimSpace(l) != "" {
			row = i
			break
		}
	}
	col = math.MaxInt32
	for _, l := range full {
		if strings.TrimSpace(l) == "" {
			continue
		}
		if n := len(l) - len(strings.TrimLeft(l, " ")); n < col {
			col = n
		}
	}
	if col == math.MaxInt32 {
		col = 0
	}
	return row, col
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Width returns the widest line, in characters.
func Width(art string) int {
	w := 0
	for _, l := range strings.Split(art, "\n") {
		if len(l) > w {
			w = len(l)
		}
	}
	return w
}
