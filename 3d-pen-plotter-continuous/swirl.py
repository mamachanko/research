#!/usr/bin/env python3
"""
Single continuous swirling stroke that suggests 3D forms.

The idea: trace a single continuous parametric curve across the surface
of a 3D object, then project to 2D. The natural foreshortening and
line density variation from perspective projection makes the brain
perceive the 3D form — even though it's just one swirly line.

No wireframes. No edges. Just one flowing stroke.
"""

import math
import xml.etree.ElementTree as ET
from typing import List, Tuple

Vec3 = Tuple[float, float, float]
Vec2 = Tuple[float, float]


# ─── Vector helpers ──────────────────────────────────────────────────────────

def v3_sub(a: Vec3, b: Vec3) -> Vec3:
    return (a[0]-b[0], a[1]-b[1], a[2]-b[2])

def v3_dot(a: Vec3, b: Vec3) -> float:
    return a[0]*b[0] + a[1]*b[1] + a[2]*b[2]

def v3_cross(a: Vec3, b: Vec3) -> Vec3:
    return (a[1]*b[2]-a[2]*b[1], a[2]*b[0]-a[0]*b[2], a[0]*b[1]-a[1]*b[0])

def v3_length(a: Vec3) -> float:
    return math.sqrt(v3_dot(a, a))

def v3_normalize(a: Vec3) -> Vec3:
    l = v3_length(a)
    return (a[0]/l, a[1]/l, a[2]/l) if l > 1e-12 else (0,0,0)

def v3_scale(a: Vec3, s: float) -> Vec3:
    return (a[0]*s, a[1]*s, a[2]*s)

def v3_add(a: Vec3, b: Vec3) -> Vec3:
    return (a[0]+b[0], a[1]+b[1], a[2]+b[2])


# ─── Projection ─────────────────────────────────────────────────────────────

def project(point: Vec3, cam_pos: Vec3, cam_target: Vec3, cam_up: Vec3,
            fov: float, width: float, height: float) -> Tuple[Vec2, float]:
    """Perspective project a 3D point. Returns (screen_xy, depth)."""
    fwd = v3_normalize(v3_sub(cam_target, cam_pos))
    right = v3_normalize(v3_cross(fwd, cam_up))
    up = v3_cross(right, fwd)

    rel = v3_sub(point, cam_pos)
    cam_x = v3_dot(rel, right)
    cam_y = v3_dot(rel, up)
    cam_z = v3_dot(rel, fwd)
    if cam_z < 0.01:
        cam_z = 0.01

    focal = 1.0 / math.tan(math.radians(fov / 2))
    scale = min(width, height) / 2 * focal
    sx = scale * (cam_x / cam_z) + width / 2
    sy = height / 2 - scale * (cam_y / cam_z)
    return (sx, sy), cam_z


# ─── Surface parametric curves ──────────────────────────────────────────────
# Each function returns a list of (x, y, z) points tracing ONE continuous
# curve across the surface. The curve is designed to "suggest" the shape
# through its density and curvature.

def sphere_spiral(radius: float = 1.0, n_wraps: int = 20, n_points: int = 2000) -> List[Vec3]:
    """
    A single spiral from south pole to north pole of a sphere.
    Dense at the equator edges, sparse at the poles — suggests roundness.
    """
    points = []
    for i in range(n_points):
        t = i / (n_points - 1)  # 0 to 1
        # latitude goes from -pi/2 to pi/2
        lat = math.pi * t - math.pi / 2
        # longitude wraps many times
        lon = 2 * math.pi * n_wraps * t

        x = radius * math.cos(lat) * math.cos(lon)
        y = radius * math.cos(lat) * math.sin(lon)
        z = radius * math.sin(lat)
        points.append((x, y, z))
    return points


def sphere_loxodrome(radius: float = 1.0, n_wraps: int = 14, n_points: int = 2000) -> List[Vec3]:
    """
    Loxodrome (rhumb line) spiral on a sphere — crosses all meridians
    at a constant angle. Creates a beautiful even spacing.
    """
    points = []
    # Loxodrome: lat = 2*atan(exp(t)) - pi/2, lon = t * slope
    slope = n_wraps
    t_range = 6.0  # controls how close to poles we get
    for i in range(n_points):
        t = -t_range + 2 * t_range * i / (n_points - 1)
        lat = 2 * math.atan(math.exp(t)) - math.pi / 2
        lon = slope * t

        x = radius * math.cos(lat) * math.cos(lon)
        y = radius * math.cos(lat) * math.sin(lon)
        z = radius * math.sin(lat)
        points.append((x, y, z))
    return points


def torus_spiral(R: float = 1.0, r: float = 0.4, n_wraps: int = 60,
                 n_points: int = 3000) -> List[Vec3]:
    """
    A single continuous spiral winding around a torus tube.
    The major angle advances slowly while the minor angle spins fast.
    """
    points = []
    for i in range(n_points):
        t = i / (n_points - 1)
        theta = 2 * math.pi * t  # major angle (around the ring)
        phi = 2 * math.pi * n_wraps * t  # minor angle (around the tube)

        x = (R + r * math.cos(phi)) * math.cos(theta)
        y = (R + r * math.cos(phi)) * math.sin(theta)
        z = r * math.sin(phi)
        points.append((x, y, z))
    return points


def cylinder_spiral(radius: float = 0.8, height: float = 2.5,
                    n_wraps: int = 16, n_points: int = 1500) -> List[Vec3]:
    """Spiral up a cylinder — suggests a cylindrical form."""
    points = []
    for i in range(n_points):
        t = i / (n_points - 1)
        angle = 2 * math.pi * n_wraps * t
        z = -height/2 + height * t
        x = radius * math.cos(angle)
        y = radius * math.sin(angle)
        points.append((x, y, z))
    return points


def cone_spiral(radius: float = 1.0, height: float = 2.0,
                n_wraps: int = 18, n_points: int = 1500) -> List[Vec3]:
    """Spiral up a cone — radius shrinks to zero at the tip."""
    points = []
    for i in range(n_points):
        t = i / (n_points - 1)
        r = radius * (1 - t)
        angle = 2 * math.pi * n_wraps * t
        z = -height/2 + height * t
        x = r * math.cos(angle)
        y = r * math.sin(angle)
        points.append((x, y, z))
    return points


def egg_spiral(a: float = 1.0, b: float = 0.7, n_wraps: int = 18,
               n_points: int = 2000) -> List[Vec3]:
    """
    Spiral on an egg/ellipsoid shape. Wider at bottom, narrower at top.
    The varying radius naturally creates density variation.
    """
    points = []
    for i in range(n_points):
        t = i / (n_points - 1)
        lat = math.pi * t - math.pi / 2
        lon = 2 * math.pi * n_wraps * t

        # Egg profile: fatter at bottom, thinner at top
        r_lat = a * math.cos(lat)
        # Asymmetry: stretch the bottom
        z = b * math.sin(lat) * (1.0 + 0.3 * math.sin(lat))

        x = r_lat * math.cos(lon)
        y = r_lat * math.sin(lon)
        points.append((x, y, z))
    return points


def moebius_spiral(radius: float = 1.0, width: float = 0.4,
                   n_cross: int = 30, n_points: int = 2000) -> List[Vec3]:
    """
    A single continuous path that weaves back and forth across
    a Möbius strip while traveling around it.
    """
    points = []
    for i in range(n_points):
        t = 2 * math.pi * i / (n_points - 1)
        half_twist = t / 2
        # Oscillate across the width of the strip
        s = width * 0.5 * math.sin(n_cross * t)

        cx = radius * math.cos(t)
        cy = radius * math.sin(t)
        x = cx + s * math.cos(half_twist) * math.cos(t)
        y = cy + s * math.cos(half_twist) * math.sin(t)
        z = s * math.sin(half_twist)
        points.append((x, y, z))
    return points


def vase_spiral(n_wraps: int = 22, n_points: int = 2000) -> List[Vec3]:
    """
    Spiral on a vase/wine-glass profile — a surface of revolution
    with a curvy silhouette.
    """
    points = []
    for i in range(n_points):
        t = i / (n_points - 1)
        z = -1.2 + 2.4 * t

        # Vase profile: bulge at bottom, narrow neck, flare at top
        r = 0.3 + 0.5 * math.exp(-3 * (z + 0.2)**2) + 0.35 * max(0, z - 0.3)**1.5

        angle = 2 * math.pi * n_wraps * t
        x = r * math.cos(angle)
        y = r * math.sin(angle)
        points.append((x, y, z))
    return points


def trefoil_knot(n_points: int = 2000, scale: float = 1.0) -> List[Vec3]:
    """
    Trefoil knot — a single continuous closed curve that naturally
    looks 3D due to its over-under crossings.
    """
    points = []
    for i in range(n_points):
        t = 2 * math.pi * i / (n_points - 1)
        x = scale * (math.sin(t) + 2 * math.sin(2*t))
        y = scale * (math.cos(t) - 2 * math.cos(2*t))
        z = scale * (-math.sin(3*t))
        points.append((x, y, z))
    return points


def spring_coil(radius: float = 0.6, coil_r: float = 0.2,
                n_coils: int = 8, n_wraps: int = 40,
                n_points: int = 2500) -> List[Vec3]:
    """
    A coil spring — the path spirals around a helical tube.
    Like a torus but stretched into a helix.
    """
    points = []
    height = 2.5
    for i in range(n_points):
        t = i / (n_points - 1)
        # Helix center
        helix_angle = 2 * math.pi * n_coils * t
        cx = radius * math.cos(helix_angle)
        cy = radius * math.sin(helix_angle)
        cz = -height/2 + height * t

        # Spiral around the helix tube
        tube_angle = 2 * math.pi * n_wraps * t
        # Local frame: radial direction and z
        rad_x = math.cos(helix_angle)
        rad_y = math.sin(helix_angle)

        x = cx + coil_r * math.cos(tube_angle) * rad_x
        y = cy + coil_r * math.cos(tube_angle) * rad_y
        z = cz + coil_r * math.sin(tube_angle)
        points.append((x, y, z))
    return points


def klein_bottle_path(n_wraps: int = 40, n_points: int = 3000,
                      scale: float = 0.8) -> List[Vec3]:
    """
    Continuous spiral path on a Klein bottle (figure-8 immersion).
    """
    points = []
    for i in range(n_points):
        t = i / (n_points - 1)
        u = 2 * math.pi * t  # major parameter
        v = 2 * math.pi * n_wraps * t  # spiral around

        # Figure-8 Klein bottle parametrization
        r = 1.0
        cos_u, sin_u = math.cos(u), math.sin(u)
        cos_v, sin_v = math.cos(v), math.sin(v)
        half_u = u / 2

        x = scale * (r + math.cos(half_u) * sin_v - math.sin(half_u) * math.sin(2*v)) * cos_u
        y = scale * (r + math.cos(half_u) * sin_v - math.sin(half_u) * math.sin(2*v)) * sin_u
        z = scale * (math.sin(half_u) * sin_v + math.cos(half_u) * math.sin(2*v))
        points.append((x, y, z))
    return points


def sphere_fibonacci(radius: float = 1.0, n_points: int = 1500) -> List[Vec3]:
    """
    Fibonacci spiral on a sphere — golden-angle spacing creates
    an organic, even distribution that suggests a sphere beautifully.
    Points are connected in order, creating a single flowing path.
    """
    golden_angle = math.pi * (3 - math.sqrt(5))  # ~137.5 degrees
    points = []
    for i in range(n_points):
        t = i / (n_points - 1)
        lat = math.asin(2 * t - 1)  # -pi/2 to pi/2
        lon = golden_angle * i

        x = radius * math.cos(lat) * math.cos(lon)
        y = radius * math.cos(lat) * math.sin(lon)
        z = radius * math.sin(lat)
        points.append((x, y, z))
    return points


# ─── SVG Rendering ──────────────────────────────────────────────────────────

def render_swirl_svg(
    points: List[Vec3],
    title: str = "",
    cam_pos: Vec3 = (3.5, 2.5, 2.0),
    cam_target: Vec3 = (0, 0, 0),
    cam_up: Vec3 = (0, 0, 1),
    fov: float = 45,
    width: float = 800,
    height: float = 800,
    stroke_color: str = "#1a1a2e",
    bg_color: str = "#faf9f6",
    stroke_min: float = 0.2,
    stroke_max: float = 2.8,
    opacity_min: float = 0.15,
    opacity_max: float = 0.95,
    smoothing: bool = True,
) -> str:
    """Render a continuous 3D curve as SVG with depth-varying stroke."""

    # Project all points
    projected = []
    depths = []
    for p in points:
        screen, depth = project(p, cam_pos, cam_target, cam_up, fov, width, height)
        projected.append(screen)
        depths.append(depth)

    z_near = min(depths)
    z_far = max(depths)
    z_range = z_far - z_near if z_far - z_near > 1e-6 else 1.0

    svg = ET.Element("svg", {
        "xmlns": "http://www.w3.org/2000/svg",
        "width": str(int(width)),
        "height": str(int(height)),
        "viewBox": f"0 0 {int(width)} {int(height)}",
    })
    ET.SubElement(svg, "rect", {"width": "100%", "height": "100%", "fill": bg_color})

    if title:
        t = ET.SubElement(svg, "text", {
            "x": str(width/2), "y": "32",
            "text-anchor": "middle",
            "font-family": "'IBM Plex Mono', monospace",
            "font-size": "15", "fill": "#555",
            "font-weight": "300",
        })
        t.text = title

    g = ET.SubElement(svg, "g", {"id": "swirl"})

    # Render as segments with per-segment depth cues
    # Group consecutive segments with similar depth into polylines for smoother rendering
    segment_batch_size = 5 if smoothing else 1

    for i in range(0, len(projected) - 1, segment_batch_size):
        end = min(i + segment_batch_size + 1, len(projected))
        batch_points = projected[i:end]
        batch_depths = depths[i:end]

        avg_depth = sum(batch_depths) / len(batch_depths)
        t_depth = (avg_depth - z_near) / z_range

        w = stroke_max - t_depth * (stroke_max - stroke_min)
        o = opacity_max - t_depth * (opacity_max - opacity_min)

        if smoothing and len(batch_points) > 2:
            # Build a smooth polyline
            pts_str = " ".join(f"{p[0]:.1f},{p[1]:.1f}" for p in batch_points)
            ET.SubElement(g, "polyline", {
                "points": pts_str,
                "fill": "none",
                "stroke": stroke_color,
                "stroke-width": f"{w:.2f}",
                "stroke-opacity": f"{o:.2f}",
                "stroke-linecap": "round",
                "stroke-linejoin": "round",
            })
        else:
            for j in range(len(batch_points) - 1):
                p0, p1 = batch_points[j], batch_points[j+1]
                ET.SubElement(g, "line", {
                    "x1": f"{p0[0]:.1f}", "y1": f"{p0[1]:.1f}",
                    "x2": f"{p1[0]:.1f}", "y2": f"{p1[1]:.1f}",
                    "stroke": stroke_color,
                    "stroke-width": f"{w:.2f}",
                    "stroke-opacity": f"{o:.2f}",
                    "stroke-linecap": "round",
                })

    return ET.tostring(svg, encoding="unicode", xml_declaration=False)


# ─── Generate all ────────────────────────────────────────────────────────────

def generate_all():
    import os
    os.chdir(os.path.dirname(os.path.abspath(__file__)))

    default_cam = dict(
        cam_pos=(3.0, 2.5, 2.0),
        cam_target=(0, 0, 0),
        cam_up=(0, 0, 1),
        fov=45,
    )

    shapes = [
        ("Sphere — Spiral", sphere_spiral(1.0, 20, 2500), {}),
        ("Sphere — Loxodrome", sphere_loxodrome(1.0, 12, 2500), {}),
        ("Sphere — Fibonacci", sphere_fibonacci(1.0, 2000), {}),
        ("Torus", torus_spiral(1.0, 0.4, 55, 4000), dict(cam_pos=(3.0, 2.0, 2.5))),
        ("Cylinder", cylinder_spiral(0.7, 2.2, 18, 1800), dict(cam_pos=(3.0, 2.0, 2.0))),
        ("Cone", cone_spiral(0.9, 2.0, 20, 1800), dict(cam_pos=(3.0, 2.0, 2.0))),
        ("Egg", egg_spiral(1.0, 0.8, 20, 2500), {}),
        ("Vase", vase_spiral(24, 2500), dict(cam_pos=(3.0, 2.0, 1.5))),
        ("Möbius Strip", moebius_spiral(1.0, 0.4, 35, 2500), dict(cam_pos=(2.5, 2.0, 1.5))),
        ("Trefoil Knot", trefoil_knot(2500, 0.5), dict(cam_pos=(4.0, 3.0, 2.5))),
        ("Spring Coil", spring_coil(0.5, 0.18, 7, 50, 3000), dict(cam_pos=(3.0, 2.0, 2.0))),
        ("Klein Bottle", klein_bottle_path(35, 3500, 0.7), dict(cam_pos=(3.5, 2.5, 2.0))),
    ]

    generated = []
    for title, pts, cam_override in shapes:
        cam = {**default_cam, **cam_override}
        fname = title.lower().replace(" ", "_").replace("—", "").replace("ö", "oe").strip("_") + ".svg"
        fname = fname.replace("__", "_")
        print(f"  {title} → {fname} ({len(pts)} points)")

        svg = render_swirl_svg(pts, title=title, **cam)
        with open(fname, "w") as f:
            f.write(svg)
        generated.append((title, fname, len(pts)))

    # Gallery HTML
    print("  → gallery.html")
    html = make_gallery(generated)
    with open("gallery.html", "w") as f:
        f.write(html)

    return generated


def make_gallery(items):
    cards = ""
    for title, fname, n_pts in items:
        cards += f"""
        <div class="card">
            <object data="{fname}" type="image/svg+xml"></object>
            <div class="caption">{title} — {n_pts:,} points, 1 stroke</div>
        </div>"""

    return f"""<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<title>Continuous Swirl — 3D Forms from One Stroke</title>
<style>
  body {{
    font-family: 'IBM Plex Mono', 'Courier New', monospace;
    background: #faf9f6; color: #1a1a2e;
    max-width: 1200px; margin: 0 auto; padding: 2rem;
  }}
  h1 {{ text-align: center; font-weight: 300; margin-bottom: 0.3rem; }}
  .subtitle {{ text-align: center; color: #888; font-size: 0.9rem; margin-bottom: 2rem; }}
  .gallery {{
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(380px, 1fr));
    gap: 1.5rem;
  }}
  .card {{
    background: white; border: 1px solid #e8e8e8; border-radius: 10px;
    padding: 1rem; text-align: center;
    box-shadow: 0 2px 12px rgba(0,0,0,0.04);
  }}
  .card object {{ width: 100%; height: auto; }}
  .caption {{ margin-top: 0.5rem; font-size: 0.8rem; color: #777; }}
</style>
</head>
<body>
  <h1>One Stroke, Three Dimensions</h1>
  <p class="subtitle">Each form drawn with a single continuous swirling line. Depth from line weight and opacity alone.</p>
  <div class="gallery">{cards}
  </div>
</body>
</html>"""


if __name__ == "__main__":
    print("Generating swirl SVGs...")
    results = generate_all()
    print(f"\nDone — {len(results)} shapes.")
