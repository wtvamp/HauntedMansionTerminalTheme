package emit_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wtvamp/HauntedMansionTerminalTheme/internal/emit"
)

// TestPortraitsFitATerminal is the guard that matters for the greeting: it
// prints one of these on every shell open and neither reflows nor truncates.
func TestPortraitsFitATerminal(t *testing.T) {
	const (
		maxCols = 80 // the narrowest terminal we promise to look right in
		maxRows = 22 // leaves the quote and a prompt inside a 24-row terminal
	)
	out, err := emit.Portraits("../../content/photos", "dist")
	if err != nil {
		t.Fatal(err)
	}
	if len(out) == 0 {
		t.Fatal("no portraits rendered")
	}
	for path, body := range out {
		name := filepath.Base(path)
		lines := strings.Split(strings.TrimRight(string(body), "\n"), "\n")
		if len(lines) > maxRows {
			t.Errorf("%s is %d rows, want <= %d — tighten its crop in render.json", name, len(lines), maxRows)
		}
		for i, line := range lines {
			if w := visibleCells(line); w > maxCols {
				t.Errorf("%s:%d is %d cells, want <= %d", name, i+1, w, maxCols)
			}
		}
		if !strings.HasSuffix(string(body), "\n") {
			t.Errorf("%s does not end in a newline", name)
		}
	}
}

func TestPortraitsAreDeterministic(t *testing.T) {
	// dist/ is committed and CI diffs it.
	a, err := emit.Portraits("../../content/photos", "dist")
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3; i++ {
		b, err := emit.Portraits("../../content/photos", "dist")
		if err != nil {
			t.Fatal(err)
		}
		for k, v := range a {
			if string(b[k]) != string(v) {
				t.Fatalf("%s renders differently between runs", k)
			}
		}
	}
}

// TestEveryPhotoIsCredited keeps the licence obligations real: these are CC
// BY-SA photographs and attribution is a condition of using them.
func TestEveryPhotoIsCredited(t *testing.T) {
	raw, err := os.ReadFile("../../content/photos/sources.json")
	if err != nil {
		t.Fatal(err)
	}
	var src map[string]struct {
		Artist, License, DescriptionURL string `json:"artist,license,descriptionurl"`
	}
	if err := json.Unmarshal(raw, &src); err != nil {
		t.Fatal(err)
	}
	credits, err := os.ReadFile("../../content/photos/ATTRIBUTION.md")
	if err != nil {
		t.Fatal(err)
	}
	photos, _ := filepath.Glob("../../content/photos/*.jpg")
	if len(photos) == 0 {
		t.Fatal("no photos found")
	}
	for _, p := range photos {
		name := strings.TrimSuffix(filepath.Base(p), ".jpg")
		if _, ok := src[name]; !ok {
			t.Errorf("%s.jpg has no entry in sources.json", name)
		}
		if !strings.Contains(string(credits), name+".jpg") {
			t.Errorf("%s.jpg is not credited in ATTRIBUTION.md", name)
		}
	}
}

// visibleCells counts printed cells, ignoring SGR sequences. Every glyph in the
// block output is one cell wide.
func visibleCells(s string) int {
	n, inEsc := 0, false
	for _, r := range s {
		switch {
		case r == 0x1b:
			inEsc = true
		case inEsc && r == 'm':
			inEsc = false
		case inEsc:
		default:
			n++
		}
	}
	return n
}
