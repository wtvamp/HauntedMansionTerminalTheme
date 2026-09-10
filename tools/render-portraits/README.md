# Portrait renderer

Draws the source images for `content/ghosts/` from primitives — original artwork,
no photographs — which `cmd/portrait` then converts to ASCII.

```sh
make portraits          # render PNGs and regenerate every content/ghosts/*.txt
```

Needs Python with Pillow (`pip install pillow`). It is a **development tool**: the
`.txt` files are committed, so nobody building or using the theme needs Python.

Editing `portraits.py` and re-running is the way to change the art. Hand-editing a
`.txt` works too, but the next `make portraits` overwrites it.

## Why the art looks the way it does

`cmd/portrait` averages each terminal cell and maps it onto a character ramp, so
at 46 columns there are only a couple of usable brightness steps per object. That
drives every choice here:

- **Busts, not scenes.** A head-and-shoulders crop fills the frame with one subject.
  A scene with several objects turns to noise — a tightrope walker over an alligator
  was tried and cut for exactly this.
- **Near-square canvas.** A tall portrait at a readable column width produces more
  rows than a 24-row terminal has.
- **One key light.** Overlapping soft gradients average out to flat mid-grey.
- **Draw hard details last.** A rim or a highlight added before the vignette gets
  erased by it — Leota's crystal ball has to be drawn after.
