// Package palette loads and validates the Haunted Mansion palette.
//
// palette/haunted-mansion.yaml is the single source of truth for every color in
// this repository. Emulator configs, the zsh prompt and the ride all resolve
// through this package; a hex literal anywhere else is a bug.
package palette

import (
	"fmt"
	"os"
	"sort"

	"gopkg.in/yaml.v3"

	palettedata "github.com/wtvamp/HauntedMansionTerminalTheme/palette"
)

// Slot is one of the sixteen ANSI color slots.
type Slot struct {
	Slot   int    `yaml:"slot"`
	Key    string `yaml:"key"`
	Name   string `yaml:"name"`
	Normal string `yaml:"normal"`
	Bright string `yaml:"bright"`
}

// UI holds the colors that are not ANSI slots.
type UI struct {
	Background          string `yaml:"background"`
	Foreground          string `yaml:"foreground"`
	Cursor              string `yaml:"cursor"`
	CursorText          string `yaml:"cursor_text"`
	SelectionBackground string `yaml:"selection_background"`
	SelectionForeground string `yaml:"selection_foreground"`
}

// Palette is the whole file.
type Palette struct {
	Name        string            `yaml:"name"`
	Slug        string            `yaml:"slug"`
	Author      string            `yaml:"author"`
	Description string            `yaml:"description"`
	UI          UI                `yaml:"ui"`
	ANSI        []Slot            `yaml:"ansi"`
	Roles       map[string]string `yaml:"roles"`

	byName map[string]Color
}

// Load reads a palette from disk. Use Default for the repo's own palette.
func Load(path string) (*Palette, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read palette: %w", err)
	}
	return parse(b)
}

// Default returns the palette compiled into the binary.
func Default() *Palette {
	p, err := parse(palettedata.YAML)
	if err != nil {
		panic(fmt.Sprintf("embedded palette is invalid, which the tests should have caught: %v", err))
	}
	return p
}

func parse(b []byte) (*Palette, error) {
	var p Palette
	if err := yaml.Unmarshal(b, &p); err != nil {
		return nil, fmt.Errorf("parse palette: %w", err)
	}
	if err := p.index(); err != nil {
		return nil, err
	}
	return &p, nil
}

// index builds the name->Color table and validates structure. Roles are resolved
// eagerly so a typo in a role alias fails at load rather than at render.
func (p *Palette) index() error {
	if len(p.ANSI) != 8 {
		return fmt.Errorf("palette: want 8 ansi entries (each with a normal and a bright), got %d", len(p.ANSI))
	}
	p.byName = make(map[string]Color, 32)

	seenSlot := map[int]bool{}
	for _, s := range p.ANSI {
		if s.Slot < 0 || s.Slot > 7 {
			return fmt.Errorf("palette: ansi %q has slot %d, want 0-7", s.Key, s.Slot)
		}
		if seenSlot[s.Slot] {
			return fmt.Errorf("palette: ansi slot %d declared twice", s.Slot)
		}
		seenSlot[s.Slot] = true

		for label, hex := range map[string]string{s.Key: s.Normal, "bright_" + s.Key: s.Bright} {
			c, err := ParseHex(hex)
			if err != nil {
				return fmt.Errorf("palette: ansi %s: %w", label, err)
			}
			p.byName[label] = c
		}
	}

	for label, hex := range map[string]string{
		"background":           p.UI.Background,
		"foreground":           p.UI.Foreground,
		"cursor":               p.UI.Cursor,
		"cursor_text":          p.UI.CursorText,
		"selection_background": p.UI.SelectionBackground,
		"selection_foreground": p.UI.SelectionForeground,
	} {
		c, err := ParseHex(hex)
		if err != nil {
			return fmt.Errorf("palette: ui %s: %w", label, err)
		}
		p.byName[label] = c
	}

	for role, target := range p.Roles {
		if _, ok := p.byName[target]; !ok {
			return fmt.Errorf("palette: role %q points at %q, which is not a color in this palette", role, target)
		}
	}
	return nil
}

// Color resolves a color by name: an ansi key ("red"), a bright key
// ("bright_red"), a ui key ("background"), or a role ("prompt_path").
func (p *Palette) Color(name string) (Color, bool) {
	if c, ok := p.byName[name]; ok {
		return c, true
	}
	if target, ok := p.Roles[name]; ok {
		c, ok := p.byName[target]
		return c, ok
	}
	return Color{}, false
}

// MustColor is Color for names the caller knows are valid. It panics otherwise,
// which is what you want in template rendering and style setup: a missing color
// is a build-time mistake, not a runtime condition to handle.
func (p *Palette) MustColor(name string) Color {
	c, ok := p.Color(name)
	if !ok {
		panic(fmt.Sprintf("palette: no color named %q", name))
	}
	return c
}

// Ordered returns the 16 ANSI colors in slot order: 0-7 normal, then 8-15 bright.
// This is the order every emulator format expects.
func (p *Palette) Ordered() []Color {
	slots := append([]Slot(nil), p.ANSI...)
	sort.Slice(slots, func(i, j int) bool { return slots[i].Slot < slots[j].Slot })

	out := make([]Color, 0, 16)
	for _, s := range slots {
		out = append(out, MustParseHex(s.Normal))
	}
	for _, s := range slots {
		out = append(out, MustParseHex(s.Bright))
	}
	return out
}

// Names returns the 16 ANSI color names in the same order as Ordered.
func (p *Palette) Names() []string {
	slots := append([]Slot(nil), p.ANSI...)
	sort.Slice(slots, func(i, j int) bool { return slots[i].Slot < slots[j].Slot })

	out := make([]string, 0, 16)
	for _, s := range slots {
		out = append(out, s.Key)
	}
	for _, s := range slots {
		out = append(out, "bright_"+s.Key)
	}
	return out
}
