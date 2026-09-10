package portrait_test

import (
	"image"
	"image/color"
	"os"
	"path/filepath"
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

// TestShippedArt guards what the greeting can actually display. The greeting
// neither reflows nor truncates, so art that is too wide is corrupt on a narrow
// terminal and art that is too tall scrolls the prompt off screen the moment a
// shell opens.
// TestEveryPortraitHasAMap keeps the two halves of a portrait together. Art
// without a map silently loses its colour; a map without art is dead weight.
func TestEveryPortraitHasAMap(t *testing.T) {
	arts, _ := filepath.Glob("../../content/ghosts/*.txt")
	maps, _ := filepath.Glob("../../content/ghosts/*.map")
	if len(arts) != len(maps) {
		t.Fatalf("%d portraits but %d colour maps; run `make portraits`", len(arts), len(maps))
	}
	for _, a := range arts {
		name := strings.TrimSuffix(a, ".txt")
		mb, err := os.ReadFile(name + ".map")
		if err != nil {
			t.Errorf("%s has no colour map", filepath.Base(a))
			continue
		}
		ab, _ := os.ReadFile(a)
		artLines := strings.Split(strings.TrimRight(string(ab), "\n"), "\n")
		mapLines := strings.Split(strings.TrimRight(string(mb), "\n"), "\n")
		if len(mapLines) != len(artLines) {
			t.Errorf("%s: art has %d rows, map has %d", filepath.Base(a), len(artLines), len(mapLines))
		}
		for _, line := range mapLines {
			for i := 0; i < len(line); i++ {
				if _, ok := portrait.RoleForKey(line[i]); !ok {
					t.Errorf("%s: map uses unknown region key %q", filepath.Base(name+".map"), line[i])
					break
				}
			}
		}
	}
}

func TestShippedArt(t *testing.T) {
	const (
		maxCols = 80 // the narrowest terminal we promise to look right in
		maxRows = 23 // leaves a prompt line in a default 24-row terminal
	)
	files, err := filepath.Glob("../../content/ghosts/*.txt")
	if err != nil || len(files) == 0 {
		t.Fatalf("no ghost art found: %v", err)
	}
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		name := filepath.Base(f)
		lines := strings.Split(strings.TrimRight(string(b), "\n"), "\n")
		if len(lines) > maxRows {
			t.Errorf("%s is %d rows, want <= %d", name, len(lines), maxRows)
		}
		for i, line := range lines {
			if len(line) > maxCols {
				t.Errorf("%s:%d is %d columns, want <= %d", name, i+1, len(line), maxCols)
			}
			for _, r := range line {
				if r < 0x20 || r > 0x7e {
					t.Errorf("%s:%d contains non-ASCII %q; the greeting cannot assume Unicode", name, i+1, r)
					break
				}
			}
		}
	}
}
