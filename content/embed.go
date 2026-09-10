// Package contentdata carries the lore files into the binaries.
//
// Same reason as palette/embed.go: go:embed cannot reach outside its directory,
// and content/ must stay the single copy. The greeting reads these files from
// disk (it has no binary); the ride reads them from here.
package contentdata

import (
	_ "embed"
	"strings"
)

//go:embed quotes.txt
var quotesRaw string

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
