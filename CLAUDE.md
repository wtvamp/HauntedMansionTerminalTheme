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
| Shell prompt + greeting | `shell/` | zsh builtins | Runs on **every** shell open — cannot spawn a runtime |
| The ride | `cmd/doombuggy/`, `internal/ride/` | Go / Bubble Tea | Static binary, zero runtime deps |
| Lore: quotes + portraits | `content/` | plain text | Contributors add lore without touching code |
| Portrait pipeline | `tools/render-portraits/`, `cmd/portrait/` | Python + Go | Dev-only; the `.txt` output is committed |

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
make portraits      # redraw content/ghosts/ (needs Python + Pillow)
make uninstall      # undo make install

go test ./internal/palette -run TestContrast    # single test
bin/doombuggy --frame graveyard                 # render one scene to stdout
bin/doombuggy --frame graveyard --at 40         # ...at a specific tick
bin/doombuggy --skip-intro                      # bypass the stretching room
bin/doombuggy --color 16                        # force a degraded palette
bin/conjure -list                               # what targets exist
bin/portrait -w 46 -ramp dense pic.png          # any image -> ASCII
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
- `content/ghosts/*.txt` — one portrait per file. **ASCII only, ≤80 columns, ≤23 rows** — the greeting
  neither reflows nor truncates, so wider art is corrupt on a narrow terminal and taller art scrolls
  the prompt away the moment a shell opens. `TestShippedArt` enforces all three. Box-drawing and
  Unicode belong in the ride, where capability is known.

  The art is **generated, not hand-drawn**: `tools/render-portraits/` draws source images from
  primitives and `cmd/portrait` converts them. `make portraits` regenerates every file. Hand-editing a
  `.txt` works until the next regeneration overwrites it — change the renderer instead. Read that
  directory's README before touching the art; the constraints there (busts not scenes, near-square
  canvas, one key light, hard details drawn last) are each a thing that was tried and failed.

Both are data — adding lore needs no code change and no rebuild. The ride reads them through
`content/embed.go`; the greeting reads them from disk.

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
