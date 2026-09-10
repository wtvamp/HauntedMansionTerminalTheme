package palette

import (
	"fmt"
	"math"
	"strings"
)

// Color is an 8-bit-per-channel sRGB color.
type Color struct{ R, G, B uint8 }

// ParseHex accepts "#RRGGBB" or "RRGGBB", case-insensitive.
func ParseHex(s string) (Color, error) {
	h := strings.TrimPrefix(strings.TrimSpace(s), "#")
	if len(h) != 6 {
		return Color{}, fmt.Errorf("color %q: want 6 hex digits, got %d", s, len(h))
	}
	var c Color
	if _, err := fmt.Sscanf(strings.ToUpper(h), "%02X%02X%02X", &c.R, &c.G, &c.B); err != nil {
		return Color{}, fmt.Errorf("color %q: %w", s, err)
	}
	return c, nil
}

func MustParseHex(s string) Color {
	c, err := ParseHex(s)
	if err != nil {
		panic(err)
	}
	return c
}

func (c Color) Hex() string { return fmt.Sprintf("#%02X%02X%02X", c.R, c.G, c.B) }

// Floats returns each channel in 0..1, the form iTerm2's plist wants.
func (c Color) Floats() (float64, float64, float64) {
	return float64(c.R) / 255, float64(c.G) / 255, float64(c.B) / 255
}

// linearize undoes the sRGB transfer function for one channel.
func linearize(v float64) float64 {
	if v <= 0.04045 {
		return v / 12.92
	}
	return math.Pow((v+0.055)/1.055, 2.4)
}

// Linear returns the color in linear-light sRGB.
func (c Color) Linear() (float64, float64, float64) {
	r, g, b := c.Floats()
	return linearize(r), linearize(g), linearize(b)
}

// Luminance is the WCAG relative luminance, 0 (black) to 1 (white).
func (c Color) Luminance() float64 {
	r, g, b := c.Linear()
	return 0.2126*r + 0.7152*g + 0.0722*b
}

// Contrast is the WCAG 2.x contrast ratio between two colors, 1:1 to 21:1.
// Order-independent.
func Contrast(a, b Color) float64 {
	la, lb := a.Luminance(), b.Luminance()
	if la < lb {
		la, lb = lb, la
	}
	return (la + 0.05) / (lb + 0.05)
}

// D65 white point.
const (
	xn = 0.95047
	yn = 1.00000
	zn = 1.08883
)

func labF(t float64) float64 {
	const d = 6.0 / 29.0
	if t > d*d*d {
		return math.Cbrt(t)
	}
	return t/(3*d*d) + 4.0/29.0
}

// Lab converts to CIE L*a*b* (D65), the space the separation checks measure in.
func (c Color) Lab() (l, a, bb float64) {
	r, g, b := c.Linear()
	x := 0.4124564*r + 0.3575761*g + 0.1804375*b
	y := 0.2126729*r + 0.7151522*g + 0.0721750*b
	z := 0.0193339*r + 0.1191920*g + 0.9503041*b

	fx, fy, fz := labF(x/xn), labF(y/yn), labF(z/zn)
	return 116*fy - 16, 500 * (fx - fy), 200 * (fy - fz)
}

// DeltaE is the CIE76 perceptual distance between two colors: plain Euclidean
// distance in Lab. CIE76 rather than CIEDE2000 on purpose — the thresholds here
// are coarse ("are these two obviously different?"), CIE76 is trivially auditable,
// and the extra accuracy of DE2000 would not move any decision this repo makes.
func DeltaE(a, b Color) float64 {
	l1, a1, b1 := a.Lab()
	l2, a2, b2 := b.Lab()
	dl, da, db := l1-l2, a1-a2, b1-b2
	return math.Sqrt(dl*dl + da*da + db*db)
}

// Vision models a form of color vision for simulation purposes.
type Vision int

const (
	Normal Vision = iota
	Protanopia
	Deuteranopia
	Tritanopia
)

func (v Vision) String() string {
	switch v {
	case Protanopia:
		return "protanopia"
	case Deuteranopia:
		return "deuteranopia"
	case Tritanopia:
		return "tritanopia"
	default:
		return "normal"
	}
}

// Machado et al. (2009) severity-1.0 transforms, applied in linear-light sRGB.
var visionMatrix = map[Vision][9]float64{
	Protanopia: {
		0.152286, 1.052583, -0.204868,
		0.114503, 0.786281, 0.099216,
		-0.003882, -0.048116, 1.051998,
	},
	Deuteranopia: {
		0.367322, 0.860646, -0.227968,
		0.280085, 0.672501, 0.047413,
		-0.011820, 0.042940, 0.968881,
	},
	Tritanopia: {
		1.255528, -0.076749, -0.178779,
		-0.078411, 0.930809, 0.147602,
		0.004733, 0.691367, 0.303900,
	},
}

func delinearize(v float64) float64 {
	if v <= 0.0031308 {
		return v * 12.92
	}
	return 1.055*math.Pow(v, 1/2.4) - 0.055
}

func clamp8(v float64) uint8 {
	switch {
	case v <= 0:
		return 0
	case v >= 1:
		return 255
	default:
		return uint8(math.Round(v * 255))
	}
}

// Simulate returns the color as seen under the given form of color vision.
func (c Color) Simulate(v Vision) Color {
	m, ok := visionMatrix[v]
	if !ok {
		return c
	}
	r, g, b := c.Linear()
	return Color{
		R: clamp8(delinearize(m[0]*r + m[1]*g + m[2]*b)),
		G: clamp8(delinearize(m[3]*r + m[4]*g + m[5]*b)),
		B: clamp8(delinearize(m[6]*r + m[7]*g + m[8]*b)),
	}
}
