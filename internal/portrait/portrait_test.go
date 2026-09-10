package portrait_test

import (
	"image"
	"image/color"
	"strings"
	"testing"

	"github.com/wtvamp/HauntedMansionTerminalTheme/internal/portrait"
)

// gradient is a left-to-right black-to-white ramp, so the converter's output is
// predictable without shipping a fixture image.
func gradient(w, h int) image.Image {
	img := image.NewGray(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.SetGray(x, y, color.Gray{Y: uint8(x * 255 / (w - 1))})
		}
	}
	return img
}

func TestConvertMapsBrightnessAcrossTheRamp(t *testing.T) {
	o := portrait.Defaults()
	o.Width, o.Ramp, o.Trim, o.Black = 10, "medium", false, 0
	art, err := portrait.Convert(gradient(100, 100), o)
	if err != nil {
		t.Fatal(err)
	}
	line := strings.Split(art, "\n")[0]
	ramp := portrait.Ramps["medium"]
	if line[0] != ramp[0] {
		t.Errorf("darkest cell is %q, want %q", line[0], ramp[0])
	}
	if line[len(line)-1] != ramp[len(ramp)-1] {
		t.Errorf("brightest cell is %q, want %q", line[len(line)-1], ramp[len(ramp)-1])
	}
	// Monotonic left to right: a gradient must not wobble.
	for i := 1; i < len(line); i++ {
		if strings.IndexByte(ramp, line[i]) < strings.IndexByte(ramp, line[i-1]) {
			t.Fatalf("ramp is not monotonic across a gradient: %q", line)
		}
	}
}

func TestInvertReversesTheRamp(t *testing.T) {
	o := portrait.Defaults()
	o.Width, o.Ramp, o.Trim, o.Black, o.Invert = 8, "medium", false, 0, true
	art, _ := portrait.Convert(gradient(80, 80), o)
	line := strings.Split(art, "\n")[0]
	ramp := portrait.Ramps["medium"]
	if line[0] != ramp[len(ramp)-1] {
		t.Errorf("inverted output starts %q, want %q", line[0], ramp[len(ramp)-1])
	}
}

func TestAspectControlsRowCount(t *testing.T) {
	o := portrait.Defaults()
	o.Width, o.Trim = 40, false
	o.Aspect = 0.5
	half, _ := portrait.Convert(gradient(100, 100), o)
	o.Aspect = 1.0
	full, _ := portrait.Convert(gradient(100, 100), o)
	if h, f := strings.Count(half, "\n"), strings.Count(full, "\n"); f < h*2-2 || f > h*2+2 {
		t.Errorf("aspect 1.0 gave %d rows, aspect 0.5 gave %d; want roughly double", f, h)
	}
}

func TestBadOptions(t *testing.T) {
	o := portrait.Defaults()
	o.Ramp = "nosuchramp"
	if _, err := portrait.Convert(gradient(10, 10), o); err == nil {
		t.Error("unknown ramp should error")
	}
	o = portrait.Defaults()
	o.Width = 0
	if _, err := portrait.Convert(gradient(10, 10), o); err == nil {
		t.Error("zero width should error")
	}
	o = portrait.Defaults()
	o.Black, o.White = 0.9, 0.1
	if _, err := portrait.Convert(gradient(10, 10), o); err == nil {
		t.Error("inverted clip points should error")
	}
}
