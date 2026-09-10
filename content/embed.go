// Package contentdata carries the lore files into the binaries.
//
// Same reason as palette/embed.go: go:embed cannot reach outside its directory,
// and content/ must stay the single copy. The greeting reads these files from
// disk (it has no binary); the ride reads them from here.
package contentdata

import (
	"embed"
	"strings"
)

//go:embed quotes.txt
var quotesRaw string

//go:embed ghosts/*.txt
var ghosts embed.FS

// Quotes returns the quote lines with comments and blanks removed, in file order.
func Quotes() []string {
	var out []string
	for _, line := range strings.Split(quotesRaw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		out = append(out, line)
	}
	return out
}

// Ghost returns one piece of art by name, without the .txt suffix.
func Ghost(name string) (string, bool) {
	b, err := ghosts.ReadFile("ghosts/" + name + ".txt")
	if err != nil {
		return "", false
	}
	return strings.TrimRight(string(b), "\n"), true
}

// GhostNames lists the available art, sorted.
func GhostNames() []string {
	entries, err := ghosts.ReadDir("ghosts")
	if err != nil {
		return nil
	}
	out := make([]string, 0, len(entries))
	for _, e := range entries {
		out = append(out, strings.TrimSuffix(e.Name(), ".txt"))
	}
	return out
}
