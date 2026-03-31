# 3D Pen Plotter — Continuous Stroke

Drawing three-dimensional forms with a single, continuous swirling stroke — not as geometrically precise wireframes, but as flowing lines that _suggest_ the shape. The line wraps around the form; perspective foreshortening bunches it at the edges and spreads it in the middle. Your brain sees a sphere, a torus, a vase — even though it's just one swirly line.

## The Idea

Imagine wrapping a single thread around a ball. As the thread spirals from pole to pole, it naturally bunches up near the edges (where the surface curves away from you) and spreads out in the middle (where the surface faces you). This density variation — a consequence of perspective projection — is what makes your brain perceive a three-dimensional sphere from a flat 2D drawing.

The same principle works for any 3D form: tori, cones, vases, knots, Klein bottles. Define a continuous parametric curve on the surface, project it with perspective, and the depth emerges from the line's own behavior.

## Depth Cues

No hidden-line removal. No shading. Just one line, with two simple depth cues:

- **Line weight**: thicker when close (2.8px), thinner when far (0.2px) — a 14:1 ratio
- **Opacity**: stronger when near (0.95), fainter when far (0.15)

These cues stack with the natural foreshortening from perspective projection to create a convincing sense of volume.

## Shapes

Twelve forms, each drawn as a single continuous stroke:

| Shape | Technique | Points |
|-------|-----------|--------|
| **Sphere — Spiral** | Pole-to-pole spiral, longitude wraps many times | 2,500 |
| **Sphere — Loxodrome** | Rhumb line crossing all meridians at constant angle | 2,500 |
| **Sphere — Fibonacci** | Golden-angle spiral, organic even distribution | 2,000 |
| **Torus** | Spiral winding around the tube while circling the ring | 4,000 |
| **Cylinder** | Helix spiraling upward | 1,800 |
| **Cone** | Helix with shrinking radius toward the tip | 1,800 |
| **Egg** | Spiral on an asymmetric ellipsoid (wider bottom) | 2,500 |
| **Vase** | Spiral on a surface of revolution with a curvy profile | 2,500 |
| **Mobius Strip** | Path oscillating across the strip width while circling | 2,500 |
| **Trefoil Knot** | The knot itself is already one continuous 3D curve | 2,500 |
| **Spring Coil** | Spiral around a helical tube — a stretched torus | 3,000 |
| **Klein Bottle** | Spiral on a figure-8 Klein bottle immersion | 3,500 |

### Three ways to swirl a sphere

The sphere is drawn three different ways to show how the _character_ of the spiral changes the feel:

- **Pole-to-pole spiral**: uniform wrapping, tight bunching at the poles, reads as a smooth globe
- **Loxodrome**: crosses every meridian at the same angle — more even spacing, slightly nautical feel
- **Fibonacci spiral**: golden-angle spacing creates an organic, almost biological quality — like a seed head or a dandelion

All three clearly read as "sphere" from the density pattern alone.

## How It Works

### 1. Parametric Surface Curves

Each shape is defined as a parametric curve `(x(t), y(t), z(t))` that traces a path across the surface of a 3D form. For a sphere spiral:

```python
lat = π * t - π/2          # south pole to north pole
lon = 2π * n_wraps * t     # wind around many times

x = r * cos(lat) * cos(lon)
y = r * cos(lat) * sin(lon)
z = r * sin(lat)
```

The key insight: the curve lives _on the surface_ of the object, so when projected, its density automatically reflects the surface's curvature.

### 2. Perspective Projection

A virtual camera projects each 3D point to 2D screen coordinates. The Z-depth (distance from camera) is retained for each point to drive the line weight and opacity.

### 3. Depth-Modulated Rendering

Each segment of the curve is drawn with stroke width and opacity proportional to its depth:

```
weight(z) = w_max - (z - z_near)/(z_far - z_near) * (w_max - w_min)
opacity(z) = o_max - (z - z_near)/(z_far - z_near) * (o_max - o_min)
```

Near parts are bold and opaque. Far parts are wispy and faint. The transition is continuous.

## Files

- **`swirl.py`** — Main implementation: 12 parametric surface curves, projection, depth-modulated SVG rendering (~400 lines, Python 3.8+ stdlib only)
- **`plotter.py`** — Alternative wireframe approach using graph Eulerization (Chinese Postman Problem + Hierholzer's algorithm) — included for comparison
- **`gallery.html`** — HTML gallery embedding all SVGs
- **`*.svg`** — Individual SVG outputs
- **`notes.md`** — Working notes

## Running

```bash
python3 swirl.py
# Generates 12 SVGs + gallery.html
```

No external dependencies — Python standard library only.

## Pen Plotter Compatibility

The SVG output is vector-only (polylines), suitable for pen plotters. For real hardware:
- Use [vpype](https://github.com/abey79/vpype) to optimize, scale, and convert to HPGL/G-code
- Line weight variation can be approximated by multi-pass (trace near/bold segments 2-3x)
- Or use a pressure-sensitive plotter that supports variable pen force
