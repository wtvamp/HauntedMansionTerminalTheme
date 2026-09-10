package palette

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// Thresholds. These are calibrated against what a dark terminal theme can
// actually achieve, not copied off a WCAG summary — see the comment on each.
const (
	// Body text. The one place full WCAG AAA is both meaningful and reachable.
	minForegroundContrast = 7.0

	// Chromatic ANSI colors are accents — prompts, ls, diff markers, log levels —
	// not body copy. 3:1 is the WCAG 1.4.11 non-text floor and is the honest
	// ceiling here: a red with 4.5:1 on a #14121A ground is no longer a red, it
	// is pink. Slots 0 and 8 are exempt and covered by minStructuralDeltaE.
	minAccentContrast = 3.0

	// Black and bright-black are structure (backgrounds, rules, dim text), so
	// contrast against the ground is inherently low. They only have to be
	// *visible* against it.
	minStructuralDeltaE = 5.0

	// Two different ANSI slots must never be mistakable for one another.
	minCrossSlotDeltaE = 15.0

	// Normal and bright of the *same* slot are meant to be the same hue at
	// different intensities, so they are held to a lower bar — but they still
	// have to be told apart, or the bright variant is decorative noise.
	minSameSlotDeltaE = 8.0

	// Red-means-error and green-means-success has to survive the two common
	// forms of red-green color blindness.
	minColorblindDeltaE = 20.0
)

func load(t *testing.T) *Palette {
	t.Helper()
	p := Default()
	if p.Slug == "" {
		t.Fatal("palette has no slug")
	}
	return p
}

func TestContrast(t *testing.T) {
	p := load(t)
	bg := p.MustColor("background")

	fg := p.MustColor("foreground")
	if got := Contrast(fg, bg); got < minForegroundContrast {
		t.Errorf("foreground %s on background %s: contrast %.2f:1, want >= %.1f:1",
			fg.Hex(), bg.Hex(), got, minForegroundContrast)
	}

	names, cols := p.Names(), p.Ordered()
	for i, name := range names {
		c := cols[i]
		if name == "black" || name == "bright_black" {
			if got := DeltaE(c, bg); got < minStructuralDeltaE {
				t.Errorf("%s %s is invisible against background %s: dE %.2f, want >= %.1f",
					name, c.Hex(), bg.Hex(), got, minStructuralDeltaE)
			}
			continue
		}
		if got := Contrast(c, bg); got < minAccentContrast {
			t.Errorf("%s %s on background: contrast %.2f:1, want >= %.1f:1",
				name, c.Hex(), got, minAccentContrast)
		}
	}

	sel := p.MustColor("selection_background")
	selFg := p.MustColor("selection_foreground")
	if got := Contrast(selFg, sel); got < minAccentContrast {
		t.Errorf("selection text %s on selection %s: contrast %.2f:1, want >= %.1f:1",
			selFg.Hex(), sel.Hex(), got, minAccentContrast)
	}

	cur := p.MustColor("cursor")
	curTxt := p.MustColor("cursor_text")
	if got := Contrast(curTxt, cur); got < minAccentContrast {
		t.Errorf("cursor text %s on cursor %s: contrast %.2f:1, want >= %.1f:1",
			curTxt.Hex(), cur.Hex(), got, minAccentContrast)
	}
}

func TestSeparation(t *testing.T) {
	p := load(t)
	names, cols := p.Names(), p.Ordered()

	sameSlot := func(a, b string) bool {
		return strings.TrimPrefix(a, "bright_") == strings.TrimPrefix(b, "bright_")
	}

	for i := 0; i < len(cols); i++ {
		for j := i + 1; j < len(cols); j++ {
			d := DeltaE(cols[i], cols[j])
			want := minCrossSlotDeltaE
			kind := "different slots"
			if sameSlot(names[i], names[j]) {
				want, kind = minSameSlotDeltaE, "same slot"
			}
			if d < want {
				t.Errorf("%s and %s (%s) are too close: dE %.2f, want >= %.1f",
					names[i], names[j], kind, d, want)
			}
		}
	}
}

func TestColorblindDistinguishable(t *testing.T) {
	p := load(t)
	for _, v := range []Vision{Protanopia, Deuteranopia} {
		for _, pair := range [][2]string{{"red", "green"}, {"bright_red", "bright_green"}} {
			a := p.MustColor(pair[0]).Simulate(v)
			b := p.MustColor(pair[1]).Simulate(v)
			if d := DeltaE(a, b); d < minColorblindDeltaE {
				t.Errorf("under %s, %s and %s collapse together: dE %.2f, want >= %.1f",
					v, pair[0], pair[1], d, minColorblindDeltaE)
			}
		}
	}
}

func TestRolesResolve(t *testing.T) {
	p := load(t)
	if len(p.Roles) == 0 {
		t.Fatal("palette declares no roles")
	}
	for role := range p.Roles {
		if _, ok := p.Color(role); !ok {
			t.Errorf("role %q does not resolve to a color", role)
		}
	}
}

// textRoles are the roles that end up rendering actual glyphs a human reads.
// They are held to the accent floor even when they point at a structural color:
// "subdued" and "invisible" are one bad alias apart, and bright_black on a crypt
// background is 1.5:1.
var textRoles = []string{
	"prompt_path", "prompt_git_clean", "prompt_git_dirty", "prompt_git_ahead",
	"prompt_error", "prompt_symbol", "greeting_ghost", "greeting_quote",
	"ride_candle", "ride_spectre", "ride_ectoplasm", "ride_velvet",
	"art_frame", "art_cloth", "art_hair", "art_skin", "art_glow", "art_accent",
}

func TestTextRolesAreLegible(t *testing.T) {
	p := load(t)
	bg := p.MustColor("background")
	for _, role := range textRoles {
		c, ok := p.Color(role)
		if !ok {
			t.Errorf("role %q is listed as a text role but is not in the palette", role)
			continue
		}
		if got := Contrast(c, bg); got < minAccentContrast {
			t.Errorf("role %s -> %s %s renders text at %.2f:1, want >= %.1f:1 (it points at a structural color)",
				role, p.Roles[role], c.Hex(), got, minAccentContrast)
		}
	}
}

func TestOrderedIsSixteenSlotsInAnsiOrder(t *testing.T) {
	p := load(t)
	cols, names := p.Ordered(), p.Names()
	if len(cols) != 16 || len(names) != 16 {
		t.Fatalf("got %d colors / %d names, want 16 of each", len(cols), len(names))
	}
	for i, want := range []string{"black", "red", "green", "yellow", "blue", "magenta", "cyan", "white"} {
		if names[i] != want {
			t.Errorf("slot %d is %q, want %q", i, names[i], want)
		}
		if names[i+8] != "bright_"+want {
			t.Errorf("slot %d is %q, want %q", i+8, names[i+8], "bright_"+want)
		}
	}
}

var hexLiteral = regexp.MustCompile(`#[0-9a-fA-F]{6}\b`)

// TestNoStrayHexLiterals enforces the rule that makes the palette a single source
// of truth: outside the palette itself and the generated output, no file may name
// a color directly. Without this the rule is a comment nobody enforces, and the
// first hand-tweaked prompt color silently forks the theme.
func TestNoStrayHexLiterals(t *testing.T) {
	root := "../.."
	skipDir := map[string]bool{
		".git": true, "dist": true, "palette": true, "bin": true, "testdata": true,
	}
	// This file names colors in its own comments, and the color math has no
	// palette to read from.
	skipFile := map[string]bool{"palette_test.go": true, "color.go": true}

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if skipDir[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		if skipFile[d.Name()] {
			return nil
		}
		switch filepath.Ext(path) {
		case ".go", ".zsh", ".sh", ".tmpl", ".yaml", ".yml":
		default:
			return nil
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for i, line := range strings.Split(string(b), "\n") {
			if m := hexLiteral.FindString(line); m != "" {
				rel, _ := filepath.Rel(root, path)
				t.Errorf("%s:%d hardcodes the color %s — add it to palette/haunted-mansion.yaml and resolve it by name instead",
					rel, i+1, m)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func ExamplePalette_Color() {
	p := Default()
	c, _ := p.Color("prompt_path")
	fmt.Println(c.Hex())
	// Output: #5FA9BE
}
