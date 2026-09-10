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
	return Entry{
		Index: index, Key: key, Label: label,
		Hex: c.Hex(), Bare: strings.TrimPrefix(c.Hex(), "#"),
		R: r, G: g, B: b,
	}
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
