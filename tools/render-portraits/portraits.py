import sys, os, math
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from PIL import Image, ImageDraw, ImageFilter
from lib import *

def constance():
    """The bride: veil, hollow gaze, hatchet at the shoulder."""
    img = canvas(); key = key_light()
    def veil(d):
        d.polygon([(W*0.16,H), (W*0.22,H*0.44), (W*0.36,H*0.10),
                   (W*0.64,H*0.10), (W*0.78,H*0.44), (W*0.84,H)], fill=255)
    lay(img, mask(veil), key.point(lambda p:int(p*0.26)), blur=26)
    box = bust(img, key, head=(0.33,0.16,0.67,0.62), shoulder_gain=0.34)
    def hair(d):
        d.ellipse([W*0.30,H*0.12,W*0.70,H*0.46], fill=255)
        d.ellipse([W*0.355,H*0.19,W*0.645,H*0.50], fill=0)
    lay(img, mask(hair), 34, blur=11)
    face(img, box, eye_y=0.42, hollow=True, mouth=0.50)
    d = ImageDraw.Draw(img)
    # hatchet over the shoulder
    d.line([(W*0.70,H*0.99),(W*0.82,H*0.70)], fill=170, width=11)
    d.polygon([(W*0.795,H*0.745),(W*0.94,H*0.66),(W*0.925,H*0.575),(W*0.775,H*0.665)], fill=225)
    return oval_frame(vignette(finish(img)))

def leota():
    """Madame Leota: a lit head floating inside a crystal ball."""
    img = canvas()
    # A tight glow only inside the ball -- a wide one flattens the whole frame
    # and the ball's own rim stops reading as an edge.
    glow = radial(W*0.50, H*0.42, W*0.34, 150, 0, falloff=2.6)
    img.paste(glow, (0,0), glow)
    key = key_light(0.42, 0.32, 0.75)
    box = bust(img, key, head=(0.37,0.22,0.63,0.56), neck=False, shoulders=None, face_gain=1.0)
    def hair(d):
        d.ellipse([W*0.335,H*0.18,W*0.665,H*0.46], fill=255)
        d.ellipse([W*0.385,H*0.235,W*0.615,H*0.50], fill=0)
    lay(img, mask(hair), 26, blur=11)
    face(img, box, eye_y=0.42, mouth=0.54, gaze=252)
    img = finish(img)
    img = vignette(img, r=0.95)
    # Ball and stand drawn last, so the vignette cannot erase the rim.
    d = ImageDraw.Draw(img)
    d.ellipse([W*0.19,H*0.11,W*0.81,H*0.73], outline=170, width=6)
    d.arc([W*0.22,H*0.14,W*0.78,H*0.70], 155, 245, fill=250, width=8)
    d.arc([W*0.22,H*0.14,W*0.78,H*0.70], 20, 70, fill=120, width=5)
    d.polygon([(W*0.34,H*0.74),(W*0.66,H*0.74),(W*0.74,H*0.90),(W*0.26,H*0.90)], fill=95)
    for i in range(7):
        x = W*(0.31+i*0.063)
        d.line([(x,H*0.75),(x-W*0.02*(i-3)*0.5,H*0.89)], fill=155, width=5)
    d.rectangle([W*0.20,H*0.90,W*0.80,H*0.97], fill=120)
    return img

def gracey():
    """Master Gracey: a gentleman, half sunk into shadow."""
    img = canvas(); key = key_light(0.30, 0.24, 1.05)
    box = bust(img, key, head=(0.34,0.15,0.66,0.58), shoulder_gain=0.30)
    def hair(d):
        d.ellipse([W*0.32,H*0.11,W*0.68,H*0.36], fill=255)
        d.ellipse([W*0.365,H*0.17,W*0.635,H*0.40], fill=0)
    lay(img, mask(hair), 45, blur=10)
    face(img, box, eye_y=0.42, mouth=0.54)
    d = ImageDraw.Draw(img)
    # cravat and lapels
    d.polygon([(W*0.50,H*0.74),(W*0.40,H*0.86),(W*0.50,H*1.0),(W*0.60,H*0.86)], fill=190)
    d.line([(W*0.36,H*0.78),(W*0.24,H*1.0)], fill=120, width=9)
    d.line([(W*0.64,H*0.78),(W*0.76,H*1.0)], fill=120, width=9)
    # the right half falls away into the dark: the portrait that ages
    shade = Image.new("L",(W,H),0)
    g = Image.new("L",(W,H))
    gd = ImageDraw.Draw(g)
    for x in range(W):
        gd.line([(x,0),(x,H)], fill=int(255*min(1.0, max(0.0,(x-W*0.52)/(W*0.42)))))
    img = Image.composite(shade, img, g.point(lambda p:int(p*0.72)))
    return oval_frame(vignette(finish(img)))

def hatbox():
    """The Hatbox Ghost: cloak, top hat, and the hatbox holding the head."""
    img = canvas(); key = key_light(0.36, 0.26, 1.1)
    def cloak(d):
        d.polygon([(W*0.50,H*0.40),(W*0.10,H*1.05),(W*0.90,H*1.05)], fill=255)
    lay(img, mask(cloak), key.point(lambda p:int(p*0.34)), blur=16)
    box = bust(img, key, head=(0.36,0.20,0.64,0.56), neck=False, shoulders=None, face_gain=0.85)
    face(img, box, eye_y=0.42, hollow=True, mouth=0.56)
    d = ImageDraw.Draw(img)
    # top hat
    d.rectangle([W*0.36,H*0.02,W*0.64,H*0.22], fill=48)
    d.ellipse([W*0.27,H*0.19,W*0.73,H*0.27], fill=62)
    d.rectangle([W*0.36,H*0.14,W*0.64,H*0.18], fill=110)
    # hatbox, held low, with a faint second face in it
    d.ellipse([W*0.60,H*0.72,W*0.98,H*0.82], fill=95)
    d.rectangle([W*0.60,H*0.77,W*0.98,H*0.98], fill=80)
    d.ellipse([W*0.60,H*0.93,W*0.98,H*1.02], fill=60)
    d.ellipse([W*0.685,H*0.83,W*0.725,H*0.865], fill=225)
    d.ellipse([W*0.855,H*0.83,W*0.895,H*0.865], fill=225)
    d.arc([W*0.70,H*0.885,W*0.88,H*0.925], 200, 340, fill=200, width=4)
    return oval_frame(vignette(finish(img)))

def raven():
    """The raven, which has been here longer than the house."""
    img = canvas(); key = key_light(0.34, 0.24, 1.1)
    def body(d):
        d.ellipse([W*0.28,H*0.30,W*0.72,H*0.78], fill=255)
        d.ellipse([W*0.52,H*0.16,W*0.78,H*0.40], fill=255)
    lay(img, mask(body), key.point(lambda p:int(p*0.55)), blur=9)
    d = ImageDraw.Draw(img)
    d.polygon([(W*0.74,H*0.26),(W*0.95,H*0.30),(W*0.74,H*0.34)], fill=180)  # beak
    d.ellipse([W*0.655,H*0.235,W*0.695,H*0.275], fill=245)                   # eye
    d.ellipse([W*0.665,H*0.245,W*0.685,H*0.265], fill=20)
    def wing(dd):
        dd.polygon([(W*0.34,H*0.38),(W*0.66,H*0.50),(W*0.40,H*0.74)], fill=255)
    lay(img, mask(wing), 60, blur=7)
    for i in range(5):
        d.arc([W*(0.34+i*0.02),H*(0.40+i*0.04),W*(0.66-i*0.03),H*(0.60+i*0.03)], 340, 60, fill=100, width=3)
    d.line([(W*0.10,H*0.86),(W*0.90,H*0.86)], fill=120, width=7)            # perch
    d.line([(W*0.46,H*0.76),(W*0.46,H*0.86)], fill=150, width=5)
    d.line([(W*0.56,H*0.76),(W*0.56,H*0.86)], fill=150, width=5)
    return vignette(finish(img), r=0.92)

def hitchhikers():
    """Ezra, Phineas and Gus, thumbs out."""
    img = canvas()
    key = key_light(0.5, 0.3, 1.3)
    for i, (cx, scale, tilt) in enumerate(((0.22,0.92,-1),(0.50,1.06,0),(0.78,0.88,1))):
        def g(d, cx=cx, s=scale):
            d.ellipse([W*(cx-0.085*s),H*(0.20),W*(cx+0.085*s),H*(0.20+0.20*s)], fill=255)
            d.polygon([(W*cx,H*(0.36*s+0.02)),(W*(cx-0.15*s),H*0.92),(W*(cx+0.15*s),H*0.92)], fill=255)
        lay(img, mask(g), key.point(lambda p:int(p*(0.62+0.12*i))), blur=10)
        d = ImageDraw.Draw(img)
        ey = H*0.28
        for ex in (W*(cx-0.045), W*(cx+0.045)):
            d.ellipse([ex-13,ey-11,ex+13,ey+11], fill=10)
            d.ellipse([ex-5,ey-4,ex+5,ey+4], fill=240)
        d.arc([W*(cx-0.05),H*0.34,W*(cx+0.05),H*0.39], 200, 340, fill=40, width=4)
        # thumb out
        d.line([(W*(cx+0.13*tilt if tilt else cx+0.13),H*0.58),
                (W*(cx+0.20*tilt if tilt else cx+0.20),H*0.50)], fill=210, width=8)
    # wavy hems
    d = ImageDraw.Draw(img)
    for i in range(28):
        x = W*(0.04+i*0.033)
        d.arc([x-16,H*0.90,x+16,H*0.96], 0, 180, fill=90, width=4)
    return vignette(finish(img), r=0.95)

def tombstone():
    """A graveyard marker, for the epitaph scenes."""
    img = canvas(); key = key_light(0.34, 0.22, 1.2)
    def stone(d):
        d.rounded_rectangle([W*0.22,H*0.18,W*0.78,H*0.88], radius=int(W*0.26), fill=255)
        d.rectangle([W*0.22,H*0.55,W*0.78,H*0.88], fill=255)
    lay(img, mask(stone), key.point(lambda p:int(p*0.62)), blur=8)
    d = ImageDraw.Draw(img)
    d.rounded_rectangle([W*0.27,H*0.24,W*0.73,H*0.82], radius=int(W*0.21), outline=40, width=5)
    # carved lettering, suggested rather than spelled
    for i, (y, wfrac) in enumerate(((0.36,0.22),(0.46,0.30),(0.54,0.26),(0.62,0.30),(0.70,0.18))):
        d.line([(W*(0.5-wfrac/2),H*y),(W*(0.5+wfrac/2),H*y)], fill=45, width=7)
    d.arc([W*0.40,H*0.26,W*0.60,H*0.34], 180, 360, fill=35, width=6)
    d.rectangle([W*0.12,H*0.86,W*0.88,H*0.96], fill=55)                # plinth
    for i in range(5):                                                  # grass
        x = W*(0.16+i*0.17)
        d.arc([x-30,H*0.90,x+30,H*1.02], 200, 340, fill=95, width=4)
    return vignette(finish(img), r=0.95)

# A full scene (a tightrope walker over an alligator) was tried and cut: at 46
# columns a scene has about two brightness steps per object and stacked soft
# gradients collapse into noise. Busts survive the downsample; scenes do not.
ALL = {
  "constance": constance, "leota": leota, "gracey": gracey, "hatbox": hatbox,
  "raven": raven, "hitchhikers": hitchhikers, "tombstone": tombstone,
}

if __name__ == "__main__":
    outdir = sys.argv[1]
    os.makedirs(outdir, exist_ok=True)
    for name, fn in ALL.items():
        save(fn(), os.path.join(outdir, name + ".png"))
        print("rendered", name)
