# Haunted Mansion — a terminal theme

999 happy haunts for your terminal: a color scheme for eight terminal emulators, a
zsh prompt, a random ghost and epitaph on every new shell, and a ride you can take
when the build is slow.

```
 ⚰ ~/Source/HauntedMansionTerminalTheme  ⑂ main ✦2 ↑1
 †
```

## Install

```sh
git clone https://github.com/wtvamp/HauntedMansionTerminalTheme.git
cd HauntedMansionTerminalTheme
make install
```

Then add to `~/.zshrc`:

```sh
export HM_ROOT=/path/to/HauntedMansionTerminalTheme
for f in ~/.zshrc.d/*.zsh(N); do source $f; done
```

And point your terminal at the matching file in `dist/`:

| Terminal | File | How |
|---|---|---|
| iTerm2 | `dist/iterm2/haunted-mansion.itermcolors` | double-click, then Preferences → Profiles → Colors → Color Presets |
| Ghostty | `dist/ghostty/haunted-mansion` | copy to `~/.config/ghostty/themes/`, then `theme = haunted-mansion` |
| Kitty | `dist/kitty/haunted-mansion.conf` | `include haunted-mansion.conf` in `kitty.conf` |
| Alacritty | `dist/alacritty/haunted-mansion.toml` | add to `import` in `alacritty.toml` |
| WezTerm | `dist/wezterm/haunted-mansion.toml` | copy to `~/.config/wezterm/colors/`, then `color_scheme = "Haunted Mansion"` |
| Windows Terminal | `dist/windows-terminal/haunted-mansion.json` | paste into the `schemes` array in `settings.json` |
| VS Code | `dist/vscode/haunted-mansion.json` | merge into `settings.json` |

You do not need Go to use the theme — `dist/` is committed.

## The ride

```sh
doombuggy                     # the full ride
doombuggy --skip-intro        # straight past the stretching room
doombuggy --frame graveyard   # one scene, one frame, to stdout
doombuggy --list              # what scenes exist
doombuggy --color 16          # force a degraded palette to check it still reads
```

`space` holds, `←`/`→` move between scenes, `q` lets you leave. Which is more than
most guests get.

## Developing

```sh
make build      # bin/conjure, bin/doombuggy
make generate   # palette -> dist/
make test       # everything
make check      # what CI runs: build, test, and verify dist/ is not stale
make demo       # run the ride without installing
```

Everything visual comes from `palette/haunted-mansion.yaml`. Change a color there
and run `make generate`; never edit anything in `dist/` by hand, and never write a
hex value into a Go file or a shell script — there is a test that fails if you do.

Adding a quote is a text file, not a code change: append a line to `content/quotes.txt`.

The portraits are **photographs of the ride**, rendered as colour half-block glyphs — real images,
not ASCII approximations of them. Sources live in `content/photos/` under free licences, credited in
[`content/photos/ATTRIBUTION.md`](content/photos/ATTRIBUTION.md); `make generate` renders them into
`dist/ghosts/`. Crop and exposure per photo live in `content/photos/render.json`.

They need a truecolor terminal. Under `NO_COLOR` the greeting prints just the quote — there is no
meaningful monochrome form of a photograph.

`portrait` works on any image, so point it at your own:

```sh
portrait -blocks -quantize=false -w 58 -gamma 1.4 -autocontrast photo.jpg
portrait -w 46 -ramp dense drawing.png     # plain ASCII, for line art
```

See `CLAUDE.md` for the architecture and the rules that govern color choices.

## Unofficial tribute

A personal fan project and an affectionate one. *The Haunted Mansion*, its characters and its script
are trademarks and copyrights of **The Walt Disney Company**. This repository is not affiliated with,
endorsed by, or sponsored by Disney. The ASCII art is drawn for this repo; the character names and
quoted lines are Disney's.
