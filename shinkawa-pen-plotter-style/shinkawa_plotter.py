"""
Shinkawa-Style Pen Plotter SVG Generator

Demonstrates techniques for translating Yoji Shinkawa's ink brush aesthetic
into pen-plotter-ready SVG files. Generates sample outputs showing:

1. Bold contour extraction with variable-width stroke simulation
2. Flow-field gestural strokes guided by image tonality
3. Dry brush texture simulation (parallel displaced lines)
4. Ink splatter / speed mark generation
5. Multi-layer composition combining all techniques

Requires: numpy, Pillow (PIL), scipy
Optional: opencv-python (for better edge detection)

Usage:
    python shinkawa_plotter.py [input_image.png]

If no input image is provided, generates a demo using synthetic shapes.
Outputs SVG files in the current directory.
"""

import math
import random
import sys
from pathlib import Path

import numpy as np
from PIL import Image, ImageFilter

# ---------------------------------------------------------------------------
# SVG helpers
# ---------------------------------------------------------------------------

SVG_HEADER = """<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg"
     width="{w}mm" height="{h}mm"
     viewBox="0 0 {w} {h}">
<g fill="none" stroke="black" stroke-linecap="round" stroke-linejoin="round">
"""
SVG_FOOTER = "</g>\n</svg>\n"


def svg_path(points, stroke_width=0.5, opacity=1.0):
    """Convert a list of (x, y) points to an SVG <path> element."""
    if len(points) < 2:
        return ""
    parts = [f"M{points[0][0]:.2f},{points[0][1]:.2f}"]
    for p in points[1:]:
        parts.append(f"L{p[0]:.2f},{p[1]:.2f}")
    style = f'stroke-width="{stroke_width:.3f}"'
    if opacity < 1.0:
        style += f' opacity="{opacity:.2f}"'
    return f'<path d="{" ".join(parts)}" {style}/>\n'


def save_svg(filename, paths_svg, w, h):
    with open(filename, "w") as f:
        f.write(SVG_HEADER.format(w=w, h=h))
        f.writelines(paths_svg)
        f.write(SVG_FOOTER)
    print(f"  Saved {filename}")


# ---------------------------------------------------------------------------
# Perlin-like noise (simple gradient noise for flow fields)
# ---------------------------------------------------------------------------

def _fade(t):
    return t * t * t * (t * (t * 6 - 15) + 10)


def _lerp(a, b, t):
    return a + t * (b - a)


class PerlinNoise2D:
    """Minimal 2D Perlin noise implementation."""

    def __init__(self, seed=42):
        rng = np.random.RandomState(seed)
        self.perm = np.arange(256, dtype=int)
        rng.shuffle(self.perm)
        self.perm = np.tile(self.perm, 2)
        angles = rng.uniform(0, 2 * math.pi, 256)
        self.grads = np.column_stack([np.cos(angles), np.sin(angles)])

    def __call__(self, x, y):
        xi, yi = int(math.floor(x)) & 255, int(math.floor(y)) & 255
        xf, yf = x - math.floor(x), y - math.floor(y)
        u, v = _fade(xf), _fade(yf)

        def grad_dot(ix, iy, dx, dy):
            g = self.grads[self.perm[self.perm[ix] + iy] & 255]
            return g[0] * dx + g[1] * dy

        n00 = grad_dot(xi, yi, xf, yf)
        n10 = grad_dot(xi + 1, yi, xf - 1, yf)
        n01 = grad_dot(xi, yi + 1, xf, yf - 1)
        n11 = grad_dot(xi + 1, yi + 1, xf - 1, yf - 1)
        return _lerp(_lerp(n00, n10, u), _lerp(n01, n11, u), v)


# ---------------------------------------------------------------------------
# Technique 1: Bold Contour Extraction
# ---------------------------------------------------------------------------

def extract_bold_contours(gray_array, threshold_high=200, threshold_low=100,
                          min_length=20):
    """
    Simple edge detection -> contour following.
    Uses Sobel-like gradient magnitude and non-maximum suppression.
    Returns list of polylines (each a list of (x,y) tuples).
    """
    from scipy.ndimage import sobel, gaussian_filter

    smooth = gaussian_filter(gray_array.astype(float), sigma=1.5)
    gx = sobel(smooth, axis=1)
    gy = sobel(smooth, axis=0)
    mag = np.sqrt(gx ** 2 + gy ** 2)

    # Normalize to 0-255
    mag = (mag / mag.max() * 255).astype(np.uint8) if mag.max() > 0 else mag

    # Threshold: keep only strong edges
    edges = mag > threshold_low

    # Trace connected edge pixels into polylines
    h, w = edges.shape
    visited = np.zeros_like(edges, dtype=bool)
    contours = []

    def neighbors(y, x):
        for dy in [-1, 0, 1]:
            for dx in [-1, 0, 1]:
                if dy == 0 and dx == 0:
                    continue
                ny, nx = y + dy, x + dx
                if 0 <= ny < h and 0 <= nx < w:
                    yield ny, nx

    for y in range(h):
        for x in range(w):
            if edges[y, x] and not visited[y, x]:
                # Trace contour
                path = [(x, y)]
                visited[y, x] = True
                cy, cx = y, x
                while True:
                    found = False
                    for ny, nx in neighbors(cy, cx):
                        if edges[ny, nx] and not visited[ny, nx]:
                            visited[ny, nx] = True
                            path.append((nx, ny))
                            cy, cx = ny, nx
                            found = True
                            break
                    if not found:
                        break
                if len(path) >= min_length:
                    contours.append(path)

    return contours


def contours_to_svg(contours, scale=1.0, base_width=0.8):
    """Convert contours to SVG paths with width variation based on length."""
    paths = []
    for contour in contours:
        # Longer contours get thicker strokes (Shinkawa's bold outlines)
        length = len(contour)
        width = base_width + min(length / 200.0, 1.5)
        scaled = [(x * scale, y * scale) for x, y in contour]
        paths.append(svg_path(scaled, stroke_width=width))
    return paths


# ---------------------------------------------------------------------------
# Technique 2: Flow Field Gestural Strokes
# ---------------------------------------------------------------------------

def generate_flow_field_strokes(gray_array, scale=1.0, num_strokes=500,
                                step_size=2.0, max_steps=80, noise_scale=0.02,
                                seed=42):
    """
    Generate gestural strokes using a Perlin noise flow field,
    weighted by image darkness (darker areas get more/longer strokes).
    """
    h, w = gray_array.shape
    noise = PerlinNoise2D(seed=seed)
    rng = random.Random(seed)

    # Invert: dark pixels -> high values (more likely to draw there)
    darkness = 1.0 - gray_array.astype(float) / 255.0

    paths = []
    attempts = 0
    max_attempts = num_strokes * 10

    while len(paths) < num_strokes and attempts < max_attempts:
        attempts += 1
        # Sample start point weighted by darkness
        sx = rng.randint(0, w - 1)
        sy = rng.randint(0, h - 1)
        if rng.random() > darkness[sy, sx] * 0.9 + 0.1:
            continue

        # Trace flow field
        points = []
        x, y = float(sx), float(sy)
        for step in range(max_steps):
            if x < 0 or x >= w or y < 0 or y >= h:
                break
            # Stop if we've entered a very light area
            ix, iy = int(x), int(y)
            if darkness[iy, ix] < 0.05 and step > 5:
                break

            points.append((x * scale, y * scale))

            # Flow direction from Perlin noise + slight image-gradient bias
            angle = noise(x * noise_scale, y * noise_scale) * math.pi * 2
            x += math.cos(angle) * step_size
            y += math.sin(angle) * step_size

        if len(points) > 5:
            # Width based on local darkness at start point
            width = 0.2 + darkness[sy, sx] * 1.0
            opacity = 0.4 + darkness[sy, sx] * 0.6
            paths.append(svg_path(points, stroke_width=width, opacity=opacity))

    return paths


# ---------------------------------------------------------------------------
# Technique 3: Dry Brush Texture Simulation
# ---------------------------------------------------------------------------

def dry_brush_stroke(points, num_bristles=5, spread=1.5, gap_prob=0.15,
                     taper_start=0.2, taper_end=0.2, seed=42):
    """
    Simulate a dry brush stroke by splitting a single path into multiple
    slightly offset parallel paths with random gaps (bristle separation).
    """
    rng = random.Random(seed)
    if len(points) < 3:
        return []

    n = len(points)
    bristle_paths = []

    for b in range(num_bristles):
        # Offset perpendicular to stroke direction
        offset = (b - num_bristles / 2) * (spread / num_bristles)
        bristle = []
        in_gap = False

        for i, (x, y) in enumerate(points):
            # Taper: reduce spread at start and end
            t = i / max(n - 1, 1)
            taper = 1.0
            if t < taper_start:
                taper = t / taper_start
            elif t > (1.0 - taper_end):
                taper = (1.0 - t) / taper_end
            taper = max(0.0, min(1.0, taper))

            # Random gaps for dry brush effect
            if rng.random() < gap_prob:
                if bristle and len(bristle) > 2:
                    bristle_paths.append(bristle)
                bristle = []
                in_gap = True
                continue
            in_gap = False

            # Compute perpendicular direction
            if i < n - 1:
                dx = points[i + 1][0] - x
                dy = points[i + 1][1] - y
            else:
                dx = x - points[i - 1][0]
                dy = y - points[i - 1][1]
            length = math.sqrt(dx * dx + dy * dy)
            if length < 0.001:
                continue
            # Perpendicular
            px, py = -dy / length, dx / length

            # Add noise to offset
            noise_amt = rng.gauss(0, spread * 0.1)
            bx = x + px * offset * taper + noise_amt
            by = y + py * offset * taper + noise_amt * 0.5
            bristle.append((bx, by))

        if len(bristle) > 2:
            bristle_paths.append(bristle)

    return bristle_paths


def generate_dry_brush_layer(gray_array, scale=1.0, num_strokes=50,
                             noise_scale=0.015, seed=42):
    """Generate dry-brush textured strokes in dark areas."""
    h, w = gray_array.shape
    noise = PerlinNoise2D(seed=seed)
    rng = random.Random(seed)
    darkness = 1.0 - gray_array.astype(float) / 255.0

    all_paths = []
    attempts = 0

    while len(all_paths) < num_strokes and attempts < num_strokes * 15:
        attempts += 1
        sx = rng.randint(10, w - 10)
        sy = rng.randint(10, h - 10)
        if darkness[sy, sx] < 0.4:
            continue

        # Generate base stroke via flow field
        points = []
        x, y = float(sx), float(sy)
        max_steps = int(40 + darkness[sy, sx] * 60)
        for _ in range(max_steps):
            if x < 2 or x >= w - 2 or y < 2 or y >= h - 2:
                break
            points.append((x * scale, y * scale))
            angle = noise(x * noise_scale, y * noise_scale) * math.pi * 2
            x += math.cos(angle) * 2.5
            y += math.sin(angle) * 2.5

        if len(points) < 8:
            continue

        # Split into bristles
        num_bristles = rng.randint(3, 7)
        spread = 1.0 + darkness[sy, sx] * 2.5
        bristles = dry_brush_stroke(points, num_bristles=num_bristles,
                                    spread=spread, gap_prob=0.12,
                                    seed=seed + attempts)
        for b in bristles:
            all_paths.append(svg_path(b, stroke_width=0.25, opacity=0.7))

    return all_paths


# ---------------------------------------------------------------------------
# Technique 4: Ink Splatter and Speed Marks
# ---------------------------------------------------------------------------

def generate_splatter(cx, cy, radius=5.0, num_dots=20, seed=42):
    """Generate ink splatter dots around a center point."""
    rng = random.Random(seed)
    paths = []
    for _ in range(num_dots):
        # Radial distribution biased toward the edge
        r = radius * (0.3 + 0.7 * rng.random())
        angle = rng.uniform(0, 2 * math.pi)
        x = cx + r * math.cos(angle)
        y = cy + r * math.sin(angle)
        # Each dot is a tiny circle approximated by a short path
        dot_size = rng.uniform(0.1, 0.4)
        dot = [(x, y), (x + dot_size, y + dot_size * 0.5)]
        paths.append(svg_path(dot, stroke_width=rng.uniform(0.2, 0.6)))
    return paths


def generate_speed_marks(x1, y1, x2, y2, num_marks=8, seed=42):
    """Generate short directional lines trailing behind a stroke direction."""
    rng = random.Random(seed)
    dx, dy = x2 - x1, y2 - y1
    length = math.sqrt(dx * dx + dy * dy)
    if length < 0.1:
        return []
    dx, dy = dx / length, dy / length

    paths = []
    for _ in range(num_marks):
        # Offset perpendicular
        perp_offset = rng.gauss(0, 3.0)
        trail_offset = rng.uniform(2, 8)
        sx = x1 - dx * trail_offset + (-dy) * perp_offset
        sy = y1 - dy * trail_offset + dx * perp_offset
        mark_len = rng.uniform(2, 6)
        ex = sx + dx * mark_len
        ey = sy + dy * mark_len
        paths.append(svg_path([(sx, sy), (ex, ey)],
                              stroke_width=rng.uniform(0.15, 0.4),
                              opacity=rng.uniform(0.3, 0.7)))
    return paths


# ---------------------------------------------------------------------------
# Technique 5: Tonal Hatching (for shadow areas)
# ---------------------------------------------------------------------------

def generate_hatching(gray_array, scale=1.0, angle_deg=45, min_spacing=1.0,
                      max_spacing=5.0, seed=42):
    """
    Generate hatching lines in dark areas of the image.
    Spacing varies inversely with darkness.
    """
    h, w = gray_array.shape
    darkness = 1.0 - gray_array.astype(float) / 255.0
    angle = math.radians(angle_deg)
    cos_a, sin_a = math.cos(angle), math.sin(angle)

    # Diagonal length of image
    diag = math.sqrt(w * w + h * h)
    paths = []

    offset = -diag
    while offset < diag:
        # Sample darkness along this hatch line to decide spacing
        line_points = []
        t = -diag
        while t < diag:
            x = w / 2 + cos_a * t - sin_a * offset
            y = h / 2 + sin_a * t + cos_a * offset
            ix, iy = int(x), int(y)
            if 0 <= ix < w and 0 <= iy < h:
                if darkness[iy, ix] > 0.3:
                    line_points.append((x * scale, y * scale))
                else:
                    if len(line_points) > 3:
                        paths.append(svg_path(line_points, stroke_width=0.3))
                    line_points = []
            t += 1.0

        if len(line_points) > 3:
            paths.append(svg_path(line_points, stroke_width=0.3))

        # Adaptive spacing: sample darkness at current offset
        sample_x = int(w / 2 - sin_a * offset)
        sample_y = int(h / 2 + cos_a * offset)
        if 0 <= sample_x < w and 0 <= sample_y < h:
            d = darkness[sample_y, sample_x]
            spacing = max_spacing - d * (max_spacing - min_spacing)
        else:
            spacing = max_spacing
        offset += spacing

    return paths


# ---------------------------------------------------------------------------
# Composition: Multi-layer Shinkawa-style output
# ---------------------------------------------------------------------------

def compose_shinkawa_style(gray_array, output_prefix="shinkawa", scale=0.5):
    """
    Generate a multi-layer SVG composition in Shinkawa's style.
    Each layer is also saved separately for multi-pen plotting.
    """
    h, w = gray_array.shape
    sw, sh = int(w * scale), int(h * scale)

    print(f"Image size: {w}x{h}, output: {sw}x{sh}mm")

    # Layer 1: Bold contours (thick brush pen)
    print("Layer 1: Bold contours...")
    contours = extract_bold_contours(gray_array, threshold_high=180,
                                     threshold_low=80, min_length=15)
    contour_svg = contours_to_svg(contours, scale=scale, base_width=0.6)
    save_svg(f"{output_prefix}_layer1_contours.svg", contour_svg, sw, sh)

    # Layer 2: Flow field gestural strokes (medium pen)
    print("Layer 2: Flow field strokes...")
    flow_svg = generate_flow_field_strokes(gray_array, scale=scale,
                                           num_strokes=300, seed=42)
    save_svg(f"{output_prefix}_layer2_flow.svg", flow_svg, sw, sh)

    # Layer 3: Dry brush texture (fine pen)
    print("Layer 3: Dry brush texture...")
    dry_svg = generate_dry_brush_layer(gray_array, scale=scale,
                                       num_strokes=40, seed=99)
    save_svg(f"{output_prefix}_layer3_drybrush.svg", dry_svg, sw, sh)

    # Layer 4: Tonal hatching (fine pen, 45 degree)
    print("Layer 4: Tonal hatching...")
    hatch_svg = generate_hatching(gray_array, scale=scale, angle_deg=45)
    save_svg(f"{output_prefix}_layer4_hatching.svg", hatch_svg, sw, sh)

    # Layer 5: Splatter accents
    print("Layer 5: Splatter accents...")
    splatter_svg = []
    darkness = 1.0 - gray_array.astype(float) / 255.0
    rng = random.Random(77)
    for _ in range(30):
        sx = rng.randint(20, w - 20)
        sy = rng.randint(20, h - 20)
        if darkness[sy, sx] > 0.5:
            splatter_svg.extend(
                generate_splatter(sx * scale, sy * scale,
                                  radius=3 + darkness[sy, sx] * 5,
                                  num_dots=int(10 + darkness[sy, sx] * 15),
                                  seed=rng.randint(0, 9999)))
    save_svg(f"{output_prefix}_layer5_splatter.svg", splatter_svg, sw, sh)

    # Combined composition
    print("Composing all layers...")
    all_svg = contour_svg + flow_svg + dry_svg + hatch_svg + splatter_svg
    save_svg(f"{output_prefix}_combined.svg", all_svg, sw, sh)

    print(f"\nDone! Generated {len(all_svg)} total paths across 5 layers.")
    print(f"Files: {output_prefix}_layer[1-5]_*.svg + {output_prefix}_combined.svg")


# ---------------------------------------------------------------------------
# Demo: generate a synthetic test image if no input provided
# ---------------------------------------------------------------------------

def create_demo_image(w=400, h=500):
    """Create a simple silhouette figure for demonstration."""
    img = Image.new("L", (w, h), 255)
    pixels = np.array(img, dtype=float)

    # Dark circle (head)
    cy_head, cx_head, r_head = 120, 200, 40
    yy, xx = np.ogrid[:h, :w]
    head_mask = ((xx - cx_head) ** 2 + (yy - cy_head) ** 2) < r_head ** 2
    pixels[head_mask] = 30

    # Torso (trapezoid approximation via gradient)
    for y in range(160, 350):
        t = (y - 160) / 190.0
        half_w = int(35 + t * 40)
        cx = 200 + int(math.sin(t * 0.5) * 10)
        left, right = max(0, cx - half_w), min(w, cx + half_w)
        darkness = 40 + t * 30
        pixels[y, left:right] = np.minimum(pixels[y, left:right], darkness)

    # Arm sweep (gestural arc)
    for t_i in range(200):
        t = t_i / 200.0
        angle = -0.8 + t * 2.2
        r = 60 + t * 80
        x = int(200 + math.cos(angle) * r)
        y = int(200 + math.sin(angle) * r)
        if 0 <= x < w and 0 <= y < h:
            thickness = int(8 * (1 - abs(t - 0.5) * 2))
            for dy in range(-thickness, thickness + 1):
                for dx in range(-thickness, thickness + 1):
                    nx, ny = x + dx, y + dy
                    if 0 <= nx < w and 0 <= ny < h:
                        pixels[ny, nx] = min(pixels[ny, nx], 50)

    # Ground shadow
    for y in range(350, 380):
        t = (y - 350) / 30.0
        half_w = int(60 * (1 - t))
        cx = 200
        left, right = max(0, cx - half_w), min(w, cx + half_w)
        darkness = 80 + t * 100
        pixels[y, left:right] = np.minimum(pixels[y, left:right], darkness)

    return np.clip(pixels, 0, 255).astype(np.uint8)


# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------

def main():
    if len(sys.argv) > 1:
        input_path = sys.argv[1]
        print(f"Loading image: {input_path}")
        img = Image.open(input_path).convert("L")
        gray = np.array(img)
        prefix = Path(input_path).stem + "_shinkawa"
    else:
        print("No input image provided. Generating demo figure...")
        gray = create_demo_image()
        prefix = "demo_shinkawa"
        # Save the demo input for reference
        Image.fromarray(gray).save(f"{prefix}_input.png")
        print(f"  Saved {prefix}_input.png")

    compose_shinkawa_style(gray, output_prefix=prefix, scale=0.5)

    print("\n--- Plotting Recommendations ---")
    print("Layer 1 (contours):  Pentel Brush Pen or thick felt-tip (1.0mm+)")
    print("Layer 2 (flow):      Medium felt-tip (0.5-0.8mm)")
    print("Layer 3 (dry brush): Worn/dry brush pen or fine felt-tip (0.3mm)")
    print("Layer 4 (hatching):  Fine technical pen (0.1-0.3mm)")
    print("Layer 5 (splatter):  Brush pen with light Z-axis contact")
    print("\nPlot layers in order 4 -> 2 -> 3 -> 1 -> 5 for best results.")
    print("Use heavyweight paper (200+ gsm) for ink absorption.")


if __name__ == "__main__":
    main()
