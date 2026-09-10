"""Procedural portrait renderer for the Haunted Mansion theme.

Draws shaded bust portraits from primitives -- original artwork, no source
photographs -- which cmd/portrait then converts to ASCII.

The canvas is close to square because that is what a head-and-shoulders crop
wants, and because a tall canvas at a readable column width produces more rows
than a 24-row terminal has.
"""
from PIL import Image, ImageDraw, ImageFilter
import math

W, H = 760, 760
BG = 0

def canvas():
    return Image.new("L", (W, H), BG)

def radial(cx, cy, r, inner=255, outer=0, falloff=1.5):
    img = Image.new("L", (W, H), outer)
    px = img.load()
    for y in range(H):
        dy = (y - cy) ** 2
        for x in range(W):
            d = math.sqrt((x - cx) ** 2 + dy)
            t = max(0.0, 1.0 - d / r)
            px[x, y] = int(outer + (inner - outer) * (t ** falloff))
    return img

def mask(fn):
    m = Image.new("L", (W, H), 0)
    fn(ImageDraw.Draw(m))
    return m

def lay(base, m, light, blur=6):
    if blur:
        m = m.filter(ImageFilter.GaussianBlur(blur))
    if isinstance(light, int):
        light = Image.new("L", (W, H), light)
    base.paste(light, (0, 0), m)
    return base

def key_light(cx=0.34, cy=0.28, r=1.15, strength=1.0):
    """Standard portrait key from the upper left."""
    return radial(W*cx, H*cy, W*r, int(255*strength), 18)

def bust(img, key, head=(0.32,0.14,0.68,0.60), neck=True, shoulders=0.70,
         face_gain=0.95, shoulder_gain=0.38):
    """Shoulders, neck and head, lit by key. Returns the head box."""
    if shoulders:
        def sh(d):
            d.ellipse([W*-0.05, H*shoulders, W*1.05, H*1.45], fill=255)
        lay(img, mask(sh), key.point(lambda p:int(p*shoulder_gain)), blur=18)
    if neck:
        def nk(d):
            d.rounded_rectangle([W*0.44,H*0.50,W*0.56,H*0.74], radius=30, fill=255)
        lay(img, mask(nk), key.point(lambda p:int(p*0.44)), blur=10)
    x0,y0,x1,y1 = head
    def hd(d):
        d.ellipse([W*x0,H*y0,W*x1,H*y1], fill=255)
    lay(img, mask(hd), key.point(lambda p:int(p*face_gain)), blur=6)
    return (W*x0, H*y0, W*x1, H*y1)

def face(img, box, eye_y=0.40, brow=True, mouth=0.52, hollow=False, gaze=235):
    """Eyes, nose and mouth placed relative to the head box."""
    d = ImageDraw.Draw(img)
    x0,y0,x1,y1 = box
    hw, hh = x1-x0, y1-y0
    cx = (x0+x1)/2
    ey = y0 + hh*eye_y
    off = hw*0.20
    r = hw*0.075
    for ex in (cx-off, cx+off):
        if hollow:
            d.ellipse([ex-r*1.5, ey-r*1.3, ex+r*1.5, ey+r*1.3], fill=6)
            d.ellipse([ex-r*0.5, ey-r*0.4, ex+r*0.5, ey+r*0.4], fill=90)
        else:
            d.ellipse([ex-r*1.4, ey-r*0.9, ex+r*1.4, ey+r*0.9], fill=24)
            d.ellipse([ex-r*0.55, ey-r*0.5, ex+r*0.55, ey+r*0.5], fill=gaze)
        if brow:
            d.arc([ex-r*1.8, ey-r*3.2, ex+r*1.8, ey-r*0.4], 200, 340, fill=40, width=max(3,int(r*0.5)))
    # nose
    d.line([(cx, ey+hh*0.04), (cx-hw*0.035, ey+hh*0.15)], fill=70, width=max(3,int(hw*0.02)))
    d.arc([cx-hw*0.06, ey+hh*0.12, cx+hw*0.02, ey+hh*0.19], 180, 360, fill=60, width=3)
    # mouth
    my = y0 + hh*mouth + hh*0.10
    d.arc([cx-hw*0.16, my-hh*0.05, cx+hw*0.16, my+hh*0.05], 200, 340, fill=55, width=max(3,int(hw*0.022)))

def vignette(img, r=0.80, floor=0):
    v = radial(W/2, H*0.48, W*r, 255, 30, falloff=1.2)
    dark = Image.new("L",(W,H),floor)
    return Image.composite(img, dark, v)

def oval_frame(img, inset=18):
    d = ImageDraw.Draw(img)
    for i, val in enumerate((60, 130, 205, 235, 205, 130, 60)):
        o = inset + i*5
        d.ellipse([o, o, W-o, H-o], outline=val, width=3)
    return img

def finish(img, blur=1.5):
    return img.filter(ImageFilter.GaussianBlur(blur))

def save(img, path):
    img.save(path); return path
