// Command conjure renders the Haunted Mansion palette into terminal config files.
//
//	conjure                       # every target into dist/
//	conjure -target ghostty       # one target
//	CONJURE_TARGET=kitty conjure  # same, via the environment
//	conjure -list                 # what can be generated
//	conjure -check                # verify dist/ matches the palette, write nothing
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/wtvamp/HauntedMansionTerminalTheme/internal/emit"
	"github.com/wtvamp/HauntedMansionTerminalTheme/internal/palette"
)

const generator = "conjure"

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "conjure:", err)
		os.Exit(1)
	}
}

func run() error {
	var (
		palettePath = flag.String("palette", "palette/haunted-mansion.yaml", "palette to render; empty uses the one built into the binary")
		outDir      = flag.String("out", "dist", "output root")
		target      = flag.String("target", os.Getenv("CONJURE_TARGET"), "comma-separated targets to render; empty renders all")
		list        = flag.Bool("list", false, "list targets and exit")
		check       = flag.Bool("check", false, "report which files are stale without writing them")
		quiet       = flag.Bool("quiet", false, "only report problems")
	)
	flag.Parse()

	if *list {
		for _, t := range emit.Targets() {
			fmt.Printf("%-18s %s\n", t.Name, t.Path)
		}
		return nil
	}

	p, err := loadPalette(*palettePath)
	if err != nil {
		return err
	}
	view := emit.NewView(p, generator)

	chosen, err := emit.Named(*target)
	if err != nil {
		return err
	}

	var stale []string
	for _, t := range chosen {
		body, err := emit.Render(t, view)
		if err != nil {
			return err
		}
		dest := filepath.Join(*outDir, t.Path)

		if *check {
			existing, err := os.ReadFile(dest)
			switch {
			case errors.Is(err, os.ErrNotExist):
				stale = append(stale, dest+" (missing)")
			case err != nil:
				return err
			case string(existing) != string(body):
				stale = append(stale, dest)
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(dest, body, 0o644); err != nil {
			return fmt.Errorf("write %s: %w", dest, err)
		}
		if !*quiet {
			fmt.Printf("  materialised %s\n", dest)
		}
	}

	if *check {
		if len(stale) > 0 {
			sort.Strings(stale)
			return fmt.Errorf("dist/ is stale — the palette changed but these were not regenerated:\n  %s\nrun: make generate",
				strings.Join(stale, "\n  "))
		}
		if !*quiet {
			fmt.Printf("dist/ matches the palette (%d targets)\n", len(chosen))
		}
	}
	return nil
}

// loadPalette prefers the file on disk so a working copy renders what the author
// is editing, and falls back to the embedded copy so the binary works anywhere.
func loadPalette(path string) (*palette.Palette, error) {
	if path == "" {
		return palette.Default(), nil
	}
	p, err := palette.Load(path)
	if errors.Is(err, os.ErrNotExist) {
		return palette.Default(), nil
	}
	if err != nil {
		return nil, err
	}
	return p, nil
}
