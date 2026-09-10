package ride

import (
	"os"
	"strings"

	"github.com/muesli/termenv"
)

// Capability is how much color the terminal can actually show. Detected once at
// startup and never re-checked: a scene that only reads correctly in truecolor is
// incomplete, because TERM=xterm over ssh is a real audience.
type Capability int

const (
	NoColor Capability = iota
	Ansi16
	Ansi256
	TrueColor
)

func (c Capability) String() string {
	switch c {
	case TrueColor:
		return "truecolor"
	case Ansi256:
		return "256-color"
	case Ansi16:
		return "16-color"
	default:
		return "no-color"
	}
}

// Detect reads the environment the way terminal programs conventionally do.
// NO_COLOR wins over everything, per https://no-color.org.
func Detect(env func(string) string) Capability {
	if env("NO_COLOR") != "" {
		return NoColor
	}
	term := strings.ToLower(env("TERM"))
	if term == "dumb" || term == "" {
		return NoColor
	}
	switch strings.ToLower(env("COLORTERM")) {
	case "truecolor", "24bit":
		return TrueColor
	}
	if strings.Contains(term, "256") {
		return Ansi256
	}
	if strings.Contains(term, "kitty") || strings.Contains(term, "ghostty") || strings.Contains(term, "alacritty") {
		return TrueColor
	}
	return Ansi16
}

// DetectEnv is Detect against the real environment.
func DetectEnv() Capability { return Detect(os.Getenv) }

// Profile maps a capability onto the termenv profile lipgloss renders with.
// This is the whole degradation mechanism: scenes always name a truecolor hex,
// and the profile quantises it down to whatever the terminal can show.
func (c Capability) Profile() termenv.Profile {
	switch c {
	case TrueColor:
		return termenv.TrueColor
	case Ansi256:
		return termenv.ANSI256
	case Ansi16:
		return termenv.ANSI
	default:
		return termenv.Ascii
	}
}
