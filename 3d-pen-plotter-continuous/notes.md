# 3D Pen Plotter Continuous Stroke — Research Notes

## Goal
Explore how to draw three-dimensional forms using a pen plotter in a single continuous stroke without lifting the pen, while maintaining the perception of depth — with artistic interpretation rather than geometric precision.

## Key Insight
You don't need wireframes. A single continuous curve that wraps around a 3D surface, when projected with perspective, naturally creates density variations that the eye reads as 3D form. Lines bunch together where the surface curves away (edges of a sphere, sides of a cylinder), and spread apart where it faces you. This is the same principle that makes a ball of yarn look round.

## Two Approaches Explored

### Approach 1: Swirling Parametric Curves (PRIMARY — what the user wanted)

The core idea: define a continuous parametric path on the surface of a 3D object, project to 2D, and let perspective foreshortening do the work.

**Why it works:**
- Surface curvature creates natural line density variation
- Perspective projection amplifies this: near surfaces spread, far surfaces compress
- Adding depth-modulated line weight (thick=near, thin=far) and opacity reinforces the effect
- The result is unmistakably 3D even though it's technically "just a swirly line"

**Implemented shapes:**
1. Sphere (3 variants: pole-to-pole spiral, loxodrome, Fibonacci)
2. Torus — spiral winding around the tube
3. Cylinder — helix
4. Cone — shrinking helix
5. Egg — spiral on asymmetric ellipsoid
6. Vase — spiral on surface of revolution
7. Möbius strip — oscillating path across the width
8. Trefoil knot — already a single continuous 3D curve
9. Spring coil — spiral around a helical tube
10. Klein bottle — spiral on figure-8 immersion

**Depth cues used:**
- Line weight: 0.2px (far) to 2.8px (near) — 14:1 ratio
- Opacity: 0.15 (far) to 0.95 (near)
- Perspective foreshortening (free from projection)

### Approach 2: Wireframe Eulerization (secondary, included for completeness)

Treats the wireframe as a graph problem: make all vertex degrees even (Chinese Postman Problem), then find Euler circuit (Hierholzer's algorithm). Mathematically elegant but produces rigid, mechanical-looking results — not the organic "swirl" quality the user described.

**Key algorithms:**
- Chinese Postman: shortest paths + minimum-weight perfect matching on odd-degree vertices
- Matching: bitmask DP for ≤22 odd vertices, greedy for larger sets
- Hierholzer's algorithm for Euler circuit in O(|E|) time
- Component bridging for disconnected graphs

## What I Learned

### About depth perception in line art
1. **Line density IS depth** — foreshortening naturally creates it; you don't have to fake it
2. **Line weight variation is the strongest single cue** — a 10:1+ thick-to-thin ratio is very effective
3. **Opacity variation stacks well with weight** — faint far-away lines recede convincingly
4. **Hidden-line removal is unnecessary** for the swirl approach — the back-of-the-sphere lines showing through actually reinforces the sense of a solid, transparent form
5. **The spiral "character" matters artistically** — Fibonacci feels organic, loxodrome feels mathematical, pole-to-pole spiral feels like a globe

### About the parametric approach
1. **Every closed surface can be spiraled** — you just need a parametric curve that covers the surface
2. **Wrap count controls density** — more wraps = finer detail but also larger SVG files
3. **Non-orientable surfaces (Möbius, Klein) work fine** — the path just follows the surface twist naturally
4. **The trefoil knot is special** — it's already a single continuous 3D curve, no surface wrapping needed
5. **Torus spiral is particularly satisfying** — the tube-winding creates a clear "donut" even at low resolution

### About pen plotter compatibility
1. SVG polylines are the natural output format for plotters
2. Line weight variation needs multi-pass on single-pen plotters (2-3 passes for bold segments)
3. Pressure-sensitive plotters could directly use the depth data
4. vpype is the standard tool for post-processing plotter SVGs

### Technical notes
- All implemented in Python stdlib only — no numpy, no matplotlib
- 2000-4000 points per curve is the sweet spot (enough detail, reasonable SVG size)
- Polyline batching (groups of 5 points) makes smoother SVGs than individual line segments
- SVG file sizes range from 75-170KB depending on point count
