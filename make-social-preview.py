#!/usr/bin/env python3
"""Generate a 1200x630 social preview card for the error-pages index site.

Reproducible: run from anywhere; fonts resolve via ERRORPAGES_FONT_DIR -> ./fonts -> ~/.fonts -> system dirs.
"""
import os
from PIL import Image, ImageDraw, ImageFont

W, H = 1200, 630
INK = (32, 32, 32)
LINK = (57, 83, 100)
BG = (255, 255, 255)
ACCENT = (92, 176, 211)
SOFT = (245, 247, 249)

def font(size, bold=False):
    name = "Montserrat-SemiBold.ttf" if bold else "Montserrat-Regular.ttf"
    for d in [os.environ.get("ERRORPAGES_FONT_DIR"), "./fonts", os.path.expanduser("~/.fonts"), "/usr/share/fonts/truetype/dejavu"]:
        if not d:
            continue
        p = os.path.join(d, name)
        if os.path.exists(p):
            return ImageFont.truetype(p, size)
    return ImageFont.truetype("/usr/share/fonts/truetype/dejavu/DejaVuSans-Bold.ttf" if bold else "/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf", size)

img = Image.new("RGB", (W, H), BG)
d = ImageDraw.Draw(img)

# Top accent bar
d.rectangle([0, 0, W, 12], fill=LINK)

# Centered wordmark
f_word = font(88, bold=True)
word = "error-pages"
w_word = d.textlength(word, font=f_word)
d.text(((W - w_word) / 2, 90), word, font=f_word, fill=INK)

# Subtitle
f_sub = font(36)
sub = "Pretty, themeable, localized HTTP error pages"
w_sub = d.textlength(sub, font=f_sub)
d.text(((W - w_sub) / 2, 220), sub, font=f_sub, fill=LINK)

# Tagline
f_tag = font(28)
tag = "Tiny zero-dep server & Docker image — drop-in for Traefik, Nginx, Kubernetes"
w_tag = d.textlength(tag, font=f_tag)
d.text(((W - w_tag) / 2, 290), tag, font=f_tag, fill=INK)

# Status code chips (centered row)
chips = [("404", "Not Found"), ("500", "Server Error"), ("503", "Unavailable")]
f_code = font(52, bold=True)
f_msg = font(26)
chip_w = 250
gap = 30
total_w = chip_w * 3 + gap * 2
x0 = (W - total_w) / 2
y0 = 380
for i, (code, msg) in enumerate(chips):
    x = x0 + i * (chip_w + gap)
    d.rounded_rectangle([x, y0, x + chip_w, y0 + 150], radius=18, fill=SOFT, outline=LINK, width=2)
    cw = d.textlength(code, font=f_code)
    d.text((x + (chip_w - cw) / 2, y0 + 22), code, font=f_code, fill=INK)
    mw = d.textlength(msg, font=f_msg)
    d.text((x + (chip_w - mw) / 2, y0 + 100), msg, font=f_msg, fill=LINK)

# URL pill (centered)
f_url = font(28)
url = "tarampampam.github.io/error-pages"
w_url = d.textlength(url, font=f_url)
pill_w = w_url + 60
pill_x = (W - pill_w) / 2
d.rounded_rectangle([pill_x, 560, pill_x + pill_w, 610], radius=25, fill=LINK)
d.text((pill_x + 30, 570), url, font=f_url, fill=(255, 255, 255))

out = os.path.join(os.path.dirname(os.path.abspath(__file__)), "social-preview.png")
img.save(out, "PNG")
print("wrote", out, os.path.getsize(out), "bytes")
