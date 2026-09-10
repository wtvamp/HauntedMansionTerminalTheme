// Package emit renders the palette into the config formats each terminal wants.
//
// Every target is a text/template over one View. Adding a terminal means adding a
// template file and one line to targets — never a new code path, and never a new
// place where a color is spelled out.
package emit

import (
	"bytes"
	"embed"
	"fmt"
	"path"
	"sort"
	"strconv"
	"strings"
	"text/template"

	"github.com/wtvamp/HauntedMansionTerminalTheme/internal/palette"
)

//go:embed templates/*.tmpl
var templates embed.FS

// Target is one output file generated from one template.
type Target struct {
	// Name is what CONJURE_TARGET matches on.
	Name string
	// Path is the output path relative to the output root.
	Path string
	// Template is the file in templates/.
	Template string
	// Comment is how this format opens a line comment, "" if it has none.
	Comment string
}

// All the formats we generate. Keep sorted by Name; the CLI lists them in order.
var targets = []Target{
	{"alacritty", "alacritty/haunted-mansion.toml", "alacritty.toml.tmpl", "#"},
	{"ghostty", "ghostty/haunted-mansion", "ghostty.conf.tmpl", "#"},
	{"iterm2", "iterm2/haunted-mansion.itermcolors", "iterm2.itermcolors.tmpl", ""},
	{"kitty", "kitty/haunted-mansion.conf", "kitty.conf.tmpl", "#"},
	{"vscode", "vscode/haunted-mansion.json", "vscode.json.tmpl", ""},
	{"wezterm", "wezterm/haunted-mansion.toml", "wezterm.toml.tmpl", "#"},
	{"windows-terminal", "windows-terminal/haunted-mansion.json", "windows-terminal.json.tmpl", ""},
	{"bat", "bat/haunted-mansion.tmTheme", "bat.tmTheme.tmpl", ""},
	{"delta", "git/haunted-mansion.gitconfig", "delta.gitconfig.tmpl", "#"},
	{"eza", "shell/_eza.zsh", "eza.zsh.tmpl", "#"},
	{"starship", "starship/haunted-mansion.toml", "starship.toml.tmpl", "#"},
	{"tools", "shell/_tools.zsh", "tools.zsh.tmpl", "#"},
	{"zsh", "shell/_colors.zsh", "zsh.tmpl", "#"},
}

// Targets returns every known target.
func Targets() []Target { return append([]Target(nil), targets...) }

// Named returns the targets matching a CONJURE_TARGET value. Empty means all.
func Named(filter string) ([]Target, error) {
	filter = strings.TrimSpace(filter)
	if filter == "" {
		return Targets(), nil
	}
	want := map[string]bool{}
	for _, n := range strings.Split(filter, ",") {
		want[strings.TrimSpace(n)] = true
	}
	var out []Target
	for _, t := range targets {
		if want[t.Name] {
			out = append(out, t)
			delete(want, t.Name)
		}
	}
	if len(want) > 0 {
		var unknown []string
		for n := range want {
			unknown = append(unknown, n)
		}
		sort.Strings(unknown)
		return nil, fmt.Errorf("unknown target(s) %s; known targets are %s",
			strings.Join(unknown, ", "), strings.Join(names(), ", "))
	}
	return out, nil
}

func names() []string {
	out := make([]string, len(targets))
	for i, t := range targets {
		out[i] = t.Name
	}
	return out
}

// Entry is one named color as a template sees it.
type Entry struct {
	Index int    // ANSI slot 0-15, -1 for non-ANSI colors
	Key   string // "red", "bright_red", "background", "prompt_path"
	Label string // human name from the palette, e.g. "oxblood velvet"
	Hex   string // hash form, e.g. "#RRGGBB"
	Bare  string // bare form for the formats that dislike a leading hash
	R     float64
	G     float64
	B     float64
	// Ansi is the SGR foreground code for this slot: 30-37 for the normal
	// colours, 90-97 for the brights. Formats like EZA_COLORS and LS_COLORS
	// speak in these rather than in hex, and using them means the output
	// follows whatever theme the terminal actually has loaded.
	Ansi string
}

// View is the whole palette flattened into what templates need.
type View struct {
	Name        string
	Slug        string
	Author      string
	Description string
	Generator   string

	ANSI  []Entry          // 16 entries, slot order
	UI    map[string]Entry // background, foreground, cursor, ...
	Roles map[string]Entry // prompt_path, ride_candle, ...

	// ITermExtras is the non-ANSI colors under the key names iTerm2's plist
	// uses. It exists because that format names things differently from every
	// other target and templates should not carry that mapping.
	ITermExtras []Entry
}

func entry(p *palette.Palette, key, label string, index int) Entry {
	c := p.MustColor(key)
	r, g, b := c.Floats()
	e := Entry{
		Index: index, Key: key, Label: label,
		Hex: c.Hex(), Bare: strings.TrimPrefix(c.Hex(), "#"),
		R: r, G: g, B: b,
	}
	switch {
	case index >= 0 && index < 8:
		e.Ansi = strconv.Itoa(30 + index)
	case index >= 8 && index < 16:
		e.Ansi = strconv.Itoa(90 + index - 8)
	}
	return e
}

// NewView flattens a palette for rendering.
func NewView(p *palette.Palette, generator string) *View {
	labels := map[string]string{}
	for _, s := range p.ANSI {
		labels[s.Key] = s.Name
		labels["bright_"+s.Key] = s.Name + " (bright)"
	}

	v := &View{
		Name: p.Name, Slug: p.Slug, Author: p.Author,
		Description: strings.TrimSpace(p.Description),
		Generator:   generator,
		UI:          map[string]Entry{},
		Roles:       map[string]Entry{},
	}
	for i, key := range p.Names() {
		v.ANSI = append(v.ANSI, entry(p, key, labels[key], i))
	}
	for _, key := range []string{
		"background", "foreground", "cursor", "cursor_text",
		"selection_background", "selection_foreground",
	} {
		v.UI[key] = entry(p, key, key, -1)
	}
	// A role resolves to an ANSI slot where it can, so the zsh prompt can emit a
	// plain slot number and stay correct on a 16-color terminal.
	slotOf := map[string]int{}
	for i, key := range p.Names() {
		slotOf[key] = i
	}
	for role, target := range p.Roles {
		e := entry(p, role, target, -1)
		if i, ok := slotOf[target]; ok {
			e.Index = i
		}
		v.Roles[role] = e
	}

	for _, m := range []struct{ label, key string }{
		{"Background Color", "background"},
		{"Foreground Color", "foreground"},
		{"Bold Color", "bright_white"},
		{"Cursor Color", "cursor"},
		{"Cursor Text Color", "cursor_text"},
		{"Selection Color", "selection_background"},
		{"Selected Text Color", "selection_foreground"},
		{"Link Color", "bright_cyan"},
		{"Badge Color", "red"},
	} {
		e := entry(p, m.key, m.key, -1)
		e.Label = m.label
		v.ITermExtras = append(v.ITermExtras, e)
	}
	return v
}

// RoleNames returns role keys in sorted order, so generated files are stable.
func (v *View) RoleNames() []string {
	out := make([]string, 0, len(v.Roles))
	for k := range v.Roles {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

var funcs = template.FuncMap{
	"upper": strings.ToUpper,
	"lower": strings.ToLower,
	// mix blends two palette colours, t=0 giving the first and t=1 the second.
	// Needed because delta wants a *background* tint for added and removed
	// lines: it has no alpha, and an 8-digit hex is silently truncated to 6,
	// which paints the line in full-saturation green and is unreadable.
	"mix": func(a, b string, t float64) (string, error) {
		ca, err := palette.ParseHex(a)
		if err != nil {
			return "", err
		}
		cb, err := palette.ParseHex(b)
		if err != nil {
			return "", err
		}
		if t < 0 {
			t = 0
		}
		if t > 1 {
			t = 1
		}
		lerp := func(x, y uint8) uint8 { return uint8(float64(x)*(1-t) + float64(y)*t) }
		return palette.Color{R: lerp(ca.R, cb.R), G: lerp(ca.G, cb.G), B: lerp(ca.B, cb.B)}.Hex(), nil
	},
	// list and dict let a template declare a table of rows inline. The bat
	// theme is thirty near-identical XML blocks; without these it is thirty
	// copies of the same twelve lines.
	"list": func(v ...interface{}) []interface{} { return v },
	"dict": func(kv ...interface{}) (map[string]interface{}, error) {
		if len(kv)%2 != 0 {
			return nil, fmt.Errorf("dict needs an even number of arguments, got %d", len(kv))
		}
		m := make(map[string]interface{}, len(kv)/2)
		for i := 0; i < len(kv); i += 2 {
			k, ok := kv[i].(string)
			if !ok {
				return nil, fmt.Errorf("dict keys must be strings, got %T", kv[i])
			}
			m[k] = kv[i+1]
		}
		return m, nil
	},
	// ansi256 maps a slot to the xterm-256 index, which is just the slot.
	"ansi256": func(i int) int { return i },
}

// Render produces the file contents for one target.
func Render(t Target, v *View) ([]byte, error) {
	tmpl, err := template.New(path.Base(t.Template)).Funcs(funcs).
		ParseFS(templates, "templates/"+t.Template)
	if err != nil {
		return nil, fmt.Errorf("parse template %s: %w", t.Template, err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, v); err != nil {
		return nil, fmt.Errorf("render %s: %w", t.Name, err)
	}
	out := buf.Bytes()
	if len(out) > 0 && out[len(out)-1] != '\n' {
		out = append(out, '\n')
	}
	return out, nil
}
