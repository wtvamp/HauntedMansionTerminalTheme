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
