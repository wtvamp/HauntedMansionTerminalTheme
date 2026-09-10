package ride

import (
	"strings"

	contentdata "github.com/wtvamp/HauntedMansionTerminalTheme/content"
	"github.com/wtvamp/HauntedMansionTerminalTheme/internal/portrait"
)

// Art renders a portrait from content/ghosts with per-region colour.
//
// It resolves the same region map the greeting uses, so a portrait looks the
// same in the ride as it does on shell start, and both follow the palette.
// Falls back to a single role when the art has no map, and to plain text at
// NoColor -- lipgloss handles the quantising between those.
func (s *Style) Art(name, fallbackRole string) (string, bool) {
	art, ok := contentdata.Ghost(name)
	if !ok {
		return "", false
	}
	cmap, hasMap := contentdata.GhostMap(name)
	if !hasMap || s.Cap == NoColor {
		return s.Paint(fallbackRole, art), true
	}

	artLines := strings.Split(art, "\n")
	mapLines := strings.Split(cmap, "\n")

	var out strings.Builder
	for i, line := range artLines {
		if i > 0 {
			out.WriteByte('\n')
		}
		var m string
		if i < len(mapLines) {
			m = mapLines[i]
		}
		// Emit one styled span per run of same-region cells. Styling per
		// character would work but multiplies the escape count by about ten.
		start, curRole := 0, ""
		flush := func(end int) {
			if end <= start {
				return
			}
			chunk := line[start:end]
			if curRole == "" {
				out.WriteString(chunk)
			} else {
				out.WriteString(s.Role(curRole).Render(chunk))
			}
		}
		for x := 0; x < len(line); x++ {
			role := ""
			if x < len(m) {
				if r, ok := portrait.RoleForKey(m[x]); ok && r != "art_ground" {
					role = r
				}
			}
			if role != curRole {
				flush(x)
				start, curRole = x, role
			}
		}
		flush(len(line))
	}
	return out.String(), true
}
