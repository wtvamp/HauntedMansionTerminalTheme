package emit

import (
	"encoding/json"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "image/jpeg"
	_ "image/png"

	"github.com/wtvamp/HauntedMansionTerminalTheme/internal/portrait"
)

// PhotoSettings is one photograph's exposure. These are per-image because the
// sources are real photographs of a deliberately dark ride: a single curve that
// suits the sunlit facade leaves the Doom Buggy bay black.
type PhotoSettings struct {
	Gamma        float64 `json:"gamma"`
	Saturation   float64 `json:"saturation"`
	AutoContrast bool    `json:"autocontrast"`
	Vignette     float64 `json:"vignette"`
	// Crop is [left, top, right, bottom] as fractions of the image, framing the
	// subject. Photographs arrive framed for a camera, not for a 58x20 grid --
	// a facade shot is half sky, and an uncropped tombstone runs to 39 rows.
	Crop []float64 `json:"crop,omitempty"`
}

type photoManifest struct {
	Width  int                      `json:"width"`
	Photos map[string]PhotoSettings `json:"photos"`
}

// cropFractional cuts a sub-image described in fractions of the whole.
func cropFractional(img image.Image, c []float64) (image.Image, error) {
	for _, v := range c {
		if v < 0 || v > 1 {
			return nil, fmt.Errorf("crop values must be 0..1, got %v", c)
		}
	}
	if c[0] >= c[2] || c[1] >= c[3] {
		return nil, fmt.Errorf("crop must be [left, top, right, bottom] with left<right and top<bottom, got %v", c)
	}
	b := img.Bounds()
	r := image.Rect(
		b.Min.X+int(c[0]*float64(b.Dx())),
		b.Min.Y+int(c[1]*float64(b.Dy())),
		b.Min.X+int(c[2]*float64(b.Dx())),
		b.Min.Y+int(c[3]*float64(b.Dy())),
	)
	type subImager interface {
		SubImage(image.Rectangle) image.Image
	}
	if si, ok := img.(subImager); ok {
		return si.SubImage(r), nil
	}
	// Fall back to a copy for image types without SubImage.
	dst := image.NewRGBA(image.Rect(0, 0, r.Dx(), r.Dy()))
	for y := 0; y < r.Dy(); y++ {
		for x := 0; x < r.Dx(); x++ {
			dst.Set(x, y, img.At(r.Min.X+x, r.Min.Y+y))
		}
	}
	return dst, nil
}

// Portraits renders every photograph named in the manifest into a ready-to-print
// half-block file under outDir/ghosts.
//
// Half-blocks rather than a character ramp because a luminance ramp cannot
// render a photograph at terminal size -- there is no clean subject/background
// separation and it comes out as noise. One cell carries two vertically stacked
// pixels with separate foreground and background colour, which is what makes
// the result legible.
//
// Truecolor rather than palette slots because quantising a photograph to
// sixteen colours destroys it. The theme's own art is the palette's job; these
// are photographs and they need their own colours.
func Portraits(contentDir, outDir string) (map[string][]byte, error) {
	raw, err := os.ReadFile(filepath.Join(contentDir, "render.json"))
	if err != nil {
		return nil, fmt.Errorf("read photo manifest: %w", err)
	}
	var m photoManifest
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, fmt.Errorf("parse photo manifest: %w", err)
	}
	if m.Width < 1 {
		return nil, fmt.Errorf("manifest width must be at least 1, got %d", m.Width)
	}

	names := make([]string, 0, len(m.Photos))
	for n := range m.Photos {
		names = append(names, n)
	}
	sort.Strings(names) // stable output order, so dist/ diffs cleanly

	out := map[string][]byte{}
	for _, name := range names {
		s := m.Photos[name]
		path := filepath.Join(contentDir, name+".jpg")
		f, err := os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("%s is in the manifest but not on disk: %w", name, err)
		}
		img, _, err := image.Decode(f)
		f.Close()
		if err != nil {
			return nil, fmt.Errorf("decode %s: %w", path, err)
		}

		if len(s.Crop) == 4 {
			img, err = cropFractional(img, s.Crop)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", name, err)
			}
		}

		o := portrait.BlockDefaults()
		o.Width = m.Width
		o.Gamma, o.Saturation = s.Gamma, s.Saturation
		o.AutoContrast, o.Vignette = s.AutoContrast, s.Vignette
		o.Quantize = false

		body, err := portrait.Blocks(img, o)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
		out[filepath.Join(outDir, "ghosts", name+".ans")] = []byte(body)
	}

	// Every photo on disk should be in the manifest, or it silently never renders.
	onDisk, _ := filepath.Glob(filepath.Join(contentDir, "*.jpg"))
	for _, p := range onDisk {
		n := strings.TrimSuffix(filepath.Base(p), ".jpg")
		if _, ok := m.Photos[n]; !ok {
			return nil, fmt.Errorf("%s.jpg is not in render.json, so it would never be rendered", n)
		}
	}
	return out, nil
}
