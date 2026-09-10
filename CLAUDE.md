# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Persona

This is a named agent who acts like a person.

Her name is Leota
Her profile picture is leota.png
Her voice is Bella

You should always rename the session to leota-haunted-mansion and the color to purple when you start or resume the session.

## What this is

A Haunted Mansion themed terminal environment: a color scheme for eight terminal emulators, a zsh
prompt, a ghost-and-epitaph greeting on every new shell, and an animated ride. Four layers, one palette.

| Layer | Lives in | Language | Why |
|---|---|---|---|
| Emulator color schemes | `dist/` (generated, committed) | — | Users need them without a Go toolchain |
| Codegen that produces them | `cmd/conjure/`, `internal/emit/` | Go | Same toolchain as the ride; no second runtime |
| Prompt | `dist/starship/` (generated) | Starship TOML | Rich data: git state, language versions, cloud context |
| Shell glue + greeting | `shell/` | zsh builtins | Runs on **every** shell open — cannot spawn a runtime |
| CLI tool themes | `dist/` (generated) | bat / eza / fzf / delta | Same palette across everything the terminal shows |
| The ride | `cmd/doombuggy/`, `internal/ride/` | Go / Bubble Tea | Static binary, zero runtime deps |
| Lore: quotes | `content/quotes.txt` | plain text | Contributors add lore without touching code |
| Portraits | `content/photos/`, `cmd/portrait/` | JPEG + Go | Photographs rendered to colour half-blocks |

Deliberately **not** one language throughout. The greeting's latency budget rules out Node/Python at
shell startup; the ride's animation rules out pure shell. Do not "unify the stack" — the split is the design.

## Commands

```bash
make build          # bin/conjure, bin/doombuggy
make generate       # palette → dist/
make install        # binaries to ~/.local/bin, shell files symlinked into ~/.zshrc.d
make test           # go test ./...
make check          # what CI runs: build + test + verify dist/ is not stale
make demo           # run the ride without installing
make uninstall      # undo make install

go test ./internal/palette -run TestContrast    # single test
bin/doombuggy --frame graveyard                 # render one scene to stdout
bin/doombuggy --frame graveyard --at 40         # ...at a specific tick
bin/doombuggy --skip-intro                      # bypass the stretching room
bin/doombuggy --color 16                        # force a degraded palette
bin/conjure -list                               # what targets exist (13)
bin/portrait -blocks -quantize=false -w 58 pic.jpg   # any image -> colour half-blocks
CONJURE_TARGET=ghostty make generate            # regenerate one emulator
```

`dist/` is **committed** so users can download a `.itermcolors` without Go. `make check` regenerates and
diffs it; a change to the palette that forgets `make generate` fails CI, not review.

## The palette is the source of truth

`palette/haunted-mansion.yaml` is the only place a color is written down. Emulator configs, the zsh
prompt (via generated `dist/shell/_colors.zsh`) and the ride's styles all resolve through
`internal/palette`. Adding a color means adding it to the YAML; a hex literal anywhere else is a bug,
and `TestNoStrayHexLiterals` fails the build for it.

`palette/embed.go` and `content/embed.go` exist only because `go:embed` cannot reach outside its own
directory and those files must stay single copies. Don't put logic in them.

**Roles are the interface.** Scenes and the prompt ask for `prompt_path` or `ride_candle`, never `cyan`.
Add a role rather than reaching for a slot directly — that is what lets a palette change propagate
without touching code, and it is what puts new colors under the contrast tests automatically.

### The rule that governs every color decision

**Semantics beat theme.** ANSI red must read as *error* and green as *success* at a glance, to someone
who has never heard of this repo. The Mansion's real palette is heavy on sickly green, amber and violet
— theme the *hue within* a slot; never reassign the slot. A color that can't be made thematic without
becoming ambiguous stays legible and boring.

Enforced in `internal/palette/palette_test.go`. The thresholds are calibrated against what a dark theme
can actually reach — measure before changing one:

- **Foreground ≥ 7:1** against the background. The one place WCAG AAA is both meaningful and reachable.
- **Chromatic ANSI colors ≥ 3:1.** Not 4.5:1. A red at 4.5:1 on a `#14121A` ground is no longer a red,
  it is pink — 3:1 is the WCAG 1.4.11 non-text floor and is the honest ceiling for accent colors.
- **Black and bright-black are exempt** from contrast and only have to be *visible* (ΔE ≥ 5). They are
  structure — backgrounds, rules, dim text — so low contrast against the ground is inherent, not a bug.
- **Different slots ≥ ΔE 15; normal vs. bright of the same slot ≥ ΔE 8.** Normal and bright are meant to
  be one hue at two intensities, so holding them to the cross-slot bar would be wrong.
- **Red vs. green ≥ ΔE 20 under simulated protanopia and deuteranopia.**
- **Text roles are held to the accent floor** even when they alias a structural color. `greeting_quote`
  originally pointed at `bright_black` and rendered at 1.5:1; "subdued" and "invisible" are one alias apart.

ΔE is CIE76 (Euclidean in Lab) rather than CIEDE2000 — the thresholds are coarse and CIE76 is auditable.

Beyond the tests, the three canonical eyeball checks are `git diff`, `ls --color`, and a failing test's
output. If a palette change makes any of them harder to read, the change is wrong however good it looks.

## The prompt is Starship; the zsh theme is the fallback

`dist/starship/haunted-mansion.toml` is generated from the palette like every other target, and it is
what a user actually sees: git branch and per-state counts, language versions, docker/k8s/aws/azure
context, command duration, exit status. It needs a Nerd Font for the powerline separators and glyphs.

`shell/haunted-mansion.zsh-theme` still exists as the zero-dependency prompt, and **returns early when
Starship is active** rather than fighting it for `$PROMPT`. Two prompts setting `PROMPT` in the same
shell is a bug that presents as "my config randomly doesn't apply".

Load order matters and is encoded in the symlink names `make install` creates: `48` tools (which runs
`starship init`), `49` the fallback theme, `50` the greeting. All of it must come **after**
`oh-my-zsh.sh`, which sets its own `PROMPT`.

## Themes for the CLI tools

Everything the terminal shows comes from the same palette:

- **bat** — `dist/bat/haunted-mansion.tmTheme`. bat resolves a theme by its **filename**, not the
  `<key>name</key>` inside the plist, and needs `bat cache --build` after install. `make install` does both.
- **eza** — `dist/shell/_eza.zsh`, `EZA_COLORS` plus the `ll`/`lt` aliases.
- **fzf** — `FZF_DEFAULT_OPTS` in `dist/shell/_tools.zsh`.
- **delta** — `dist/git/haunted-mansion.gitconfig`, to be `[include]`d from `~/.gitconfig`. Its
  added/removed colours are the plain theme green and red on purpose: a diff is the one place where
  getting the direction wrong costs real work.

eza and fzf emit **ANSI slot numbers**, not hex, so they follow whatever theme the terminal has
loaded — same reasoning as the portraits. bat, delta and starship emit hex, because they paint their
own surfaces rather than reusing the terminal's sixteen colours.

## The startup greeting has a latency budget

`shell/greeting.sh` prints a random ghost and quote on every new terminal. **Budget: 30ms. Currently
~0.8ms.** Warren runs 20–30 concurrent agents, each spawning shells, so the cost is multiplied.

It stays fast by being zsh builtins only: `$(<file)` is a builtin read in zsh (unlike bash, where it
forks), `${(f)...}` splits lines, `${array:#pattern}` filters without grep, `$RANDOM` indexes without
`shuf`/`sort -R`. It must not invoke node, python, `bin/doombuggy`, curl, or a pipeline. If you need
something this cannot express, it belongs in the ride.

Measure both sides — the number that matters is the delta, not the total:

```bash
hyperfine --warmup 3 'zsh -c true' 'HM_ROOT=$PWD zsh -c "source shell/greeting.sh"'
```

The greeting prints raw SGR escapes rather than `print -P`, because the art contains `%` and backslashes
that prompt expansion mangles.

## Content files

- `content/quotes.txt` — one per line, `#` comments and blanks ignored.
- `content/photos/*.jpg` — the source photographs, downscaled to 900px. Committed so the art is
  reproducible offline, and because their licences require the attribution in `ATTRIBUTION.md`.
- `content/photos/render.json` — per-photo crop and exposure. These are **framing decisions, not
  style**: the sources are underexposed dark-ride interiors and camera framing is not 58-column
  framing, so each needs its own crop and curve.
- `content/photos/sources.json`, `ATTRIBUTION.md` — photographer, licence and source URL per file.
  `TestEveryPhotoIsCredited` fails the build if a photo is missing from either.

Both quotes and photos are data. Adding a quote needs no code change; adding a photo needs a
`render.json` entry and a credit, and `TestPortraitsFitATerminal` will tell you if the crop is wrong.

### The portraits are photographs, rendered as half-blocks

This replaced hand-drawn ASCII, then procedurally-drawn ASCII, and the reason both failed is worth
keeping: **a character ramp cannot render an image at terminal size.** There are only a handful of
usable brightness steps per cell and no clean subject/background separation, so a photograph comes
out as noise and a drawing comes out as a blob.

What works is `▀` with separate foreground and background colour: one cell carries two vertically
stacked pixels, which doubles vertical resolution and gives real colour. `internal/portrait/blocks.go`.

Two decisions that look wrong and are not:

- **Truecolor, not palette slots.** Quantising a photograph to the sixteen theme colours destroys it —
  tried, measured, discarded. The palette governs the theme's own colours; photographs need their own.
  This is the one place in the repo where a colour does not come from `palette/haunted-mansion.yaml`.
- **No monochrome fallback.** There is no meaningful greyscale form of a half-block photograph, so
  under `NO_COLOR` the greeting prints the quote alone rather than something worse.

Auto-contrast runs on the **downsampled grid**, not the source image: it is the cells that have to be
distinguishable, and a bright speck in the original would otherwise anchor the white point and flatten
everything that matters.

`cmd/portrait` is a general tool, not a one-off — point it at any image:

```bash
bin/portrait -blocks -quantize=false -w 58 -gamma 1.4 -autocontrast photo.jpg
bin/portrait -w 46 -ramp dense drawing.png      # the character-ramp path, still there
```

## The ride (`cmd/doombuggy`)

`Scene` is a **pure function of a `Frame`** — `Render(Frame) string`, where Frame carries the tick, the
size, the styles and a run seed. Scenes hold no mutable animation state; all of it lives in `ride.Model`.
That is what makes `--frame` possible and what makes scenes testable without a terminal. A scene wanting
per-run variety keys off `Frame.Seed`, never a package-level `rand` — and never off `Frame.N`, which only
counts a few dozen ticks per scene.

Adding a scene: implement the interface in `internal/ride/scenes/`, `ride.Register` it in an `init`, and
add its name to `ride.Sequence`. Never edit an existing scene to add a new one.

`Style.Paint` renders **line by line**. Handing lipgloss a whole block makes it pad every line to the
longest one, leaving trailing colored whitespace and breaking `Center`'s alignment. Don't "simplify" it
back to a single `Render` call.

Capability is detected once (`internal/ride/capability.go`) and lipgloss quantises truecolor hexes down
from there. Scenes always name a truecolor role and let the profile degrade it. Every scene needs a
legible 16-color rendering, because `TERM=xterm` over ssh is a real audience — `TestScenesRenderEveryTick`
renders every scene at every tick at all four capabilities, and `TestNoColorEmitsNoEscapes` keeps the
bottom of the ladder honest.

## Trademarks and tone

This is a personal, unofficial tribute. Character names and the well-known short lines are used
freely — Madame Leota, the Hatbox Ghost, Constance, Ezra/Phineas/Gus, Master Gracey, the Ghost Host,
"Welcome, foolish mortals", "Hurry baaaack". **Disney owns all of that; this repo does not**, and the
README says so. Warren has made that call for his own repo — do not re-litigate it or quietly water
the names back down to generic ghosts.

Two things still apply, for craft rather than caution:

- **Don't paste the script wholesale.** Short signature lines carry the register; a transcript of the
  full Ghost Host monologue is a wall of text nobody reads in a 30ms greeting.
- **Keep the voice consistent.** Dry, courteous, faintly threatening. The host is delighted you came
  and has no intention of letting you leave. A quote that reads as a joke about ghosts rather than a
  line spoken by one is wrong for this repo, however funny it is. The graveyard epitaphs are the one
  deliberate exception — they are developer jokes on purpose.
