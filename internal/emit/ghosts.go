package emit

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/wtvamp/HauntedMansionTerminalTheme/internal/palette"
	"github.com/wtvamp/HauntedMansionTerminalTheme/internal/portrait"
)

// Ghosts renders the portraits in contentDir into ready-to-print coloured files
// under outDir/ghosts.
//
// The colours are emitted as ANSI slot numbers rather than hex, which is what
// makes this degrade for free: the terminal resolves slot 14 using whatever
// theme is loaded, so the art is correct at 16 colours, at 256, and in
// truecolor, and a user running a different scheme gets their own palette.
//
// Baking the escapes at generate time is deliberate. The greeting has a 30ms
// budget and colouring a thousand cells in zsh at every shell start would spend
// all of it; printing a finished file costs one read.
func Ghosts(p *palette.Palette, contentDir, outDir string) (map[string][]byte, error) {
	arts, err := filepath.Glob(filepath.Join(contentDir, "*.txt"))
	if err != nil {
		return nil, err
	}
	if len(arts) == 0 {
		return nil, fmt.Errorf("no portraits in %s", contentDir)
	}

	dest := filepath.Join(outDir, "ghosts")
	out := map[string][]byte{}
	for _, artPath := range arts {
		name := strings.TrimSuffix(filepath.Base(artPath), ".txt")
		art, err := os.ReadFile(artPath)
		if err != nil {
			return nil, err
		}
		cmap, err := os.ReadFile(filepath.Join(contentDir, name+".map"))
		if err != nil {
			return nil, fmt.Errorf("%s has no colour map (%s); run `make portraits`", name, err)
		}

		body, err := colorize(p, string(art), string(cmap))
		if err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
		out[filepath.Join(dest, name+".ans")] = []byte(body)
	}
	return out, nil
}

const reset = "\x1b[0m"

// colorize pairs art with its map, emitting one escape per run of same-coloured
// cells rather than one per character -- the difference is roughly a 10x smaller
// file for identical output.
func colorize(p *palette.Palette, art, cmap string) (string, error) {
	artLines := strings.Split(strings.TrimRight(art, "\n"), "\n")
	mapLines := strings.Split(strings.TrimRight(cmap, "\n"), "\n")
	if len(mapLines) < len(artLines) {
		return "", fmt.Errorf("colour map has %d rows, art has %d", len(mapLines), len(artLines))
	}

	slots := map[byte]int{}
	for _, r := range portrait.REGIONS {
		if r.Role == "art_ground" {
			continue
		}
		slot, ok := p.SlotOf(r.Role)
		if !ok {
			return "", fmt.Errorf("role %q does not resolve to an ANSI slot", r.Role)
		}
		slots[r.Key] = slot
	}

	var sb strings.Builder
	for i, line := range artLines {
		m := mapLines[i]
		cur := -1
		for x := 0; x < len(line); x++ {
			var key byte = '.'
			if x < len(m) {
				key = m[x]
			}
			slot, coloured := slots[key]
			if !coloured {
				slot = -1
			}
			if slot != cur {
				if slot < 0 {
					sb.WriteString(reset)
				} else {
					fmt.Fprintf(&sb, "\x1b[38;5;%dm", slot)
				}
				cur = slot
			}
			sb.WriteByte(line[x])
		}
		sb.WriteString(reset)
		sb.WriteByte('\n')
	}
	return sb.String(), nil
}
