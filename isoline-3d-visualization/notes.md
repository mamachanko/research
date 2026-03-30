# Isoline 3D Visualization — Research Notes

## Goal
Explore different approaches to visualizing three-dimensional objects using isolines (contour lines) only. Build working demonstrations of each technique.

## Planned Approaches
1. **Horizontal slicing** — classic topographic contour maps, z-level slicing
2. **Joy Division / ridgeline** — stacked cross-section profiles
3. **Rotating isoline animation** — animated contour views from different angles
4. **Surface-projected isolines** — lines of constant value on parametric surfaces
5. **Depth-layered contours** — opacity/thickness encodes depth
6. **Cross-hatched isolines** — overlapping contour sets from multiple axes
7. **Radial isolines** — distance-from-center contours for spherical objects
8. **Animated morphing** — smooth transitions between contour sets

## Tools
- Python 3 with matplotlib for contour generation and 3D projection
- NumPy for surface computation
- Pillow for GIF assembly
- PNG output for static images, GIF for animations

## Implementation Notes

### Horizontal Slicing (01)
- Used matplotlib's `contour` and `contourf` with dark background
- Three surfaces: MATLAB peaks, Himmelblau, saddle
- `contourf` at low alpha provides subtle fill that aids readability without violating "lines only" spirit
- Cividis, inferno, viridis colormaps work well on dark backgrounds

### Joy Division (02)
- Key insight: back-to-front rendering with background-colored `fill_between` creates occlusion
- Vertical spacing and amplitude scale require manual tuning per surface
- Alpha gradient (faint at back, bright at front) adds depth
- The ripple function (sin(r)/r) produces a particularly striking result

### Rotating Isolines (03)
- Used matplotlib's 3D projection with `view_init(elev, azim)` varying per frame
- Drew iso-parameter curves (constant-u, constant-v) rather than height contours
- `set_pane_color` was removed in newer matplotlib — used `set_facecolor` instead
- 36 frames at 80ms each produces a smooth 2.9s rotation loop

### Depth-Layered (04)
- Line thickness range 0.3–2.0pt and alpha range 0.15–0.85 gave good depth separation
- The spiral peak surface (exp(-r) * cos(r - 2*theta)) was an unexpected highlight
- This is the simplest technique that adds depth to standard contours

### Cross-Hatched (05)
- Overlaying Z, Z+0.8X, and Z+0.8Y contours creates a hatching pattern
- The egg carton function (sin(x)*cos(y)) produces beautiful periodic cross-hatching
- Three distinct colors (red, blue, green) keep the three axes distinguishable

### Radial Isolines (06)
- Sphere: latitude circles at regular polar angles + 12 meridian great circles
- Torus: 24 toroidal circles + 12 poloidal circles
- The flat projection flattens depth but the line pattern still reads as 3D
- Color-coding by height (cool colormap) helps for the sphere

### Parametric Surfaces (07)
- Klein bottle (figure-8 immersion), Mobius strip, trefoil knot tube
- Two-color scheme (one for constant-u, another for constant-v) clarifies topology
- 3D projection at fixed angle — these surfaces really benefit from rotation too
- The trefoil knot tube needed a Frenet frame approximation for the cross-section

### Morphing Contours (08)
- Smooth hermite interpolation: t² * (3 - 2t) for smooth-step easing
- Four surfaces in a loop: single peak → four peaks → ring → saddle → single peak
- Topological transitions (contour splitting/merging) are clearly visible
- 15 frames per transition, 4 transitions = 60 frames total at 100ms

## Observations
- Occlusion (Joy Division style) gives the strongest immediate 3D impression
- Line density naturally encodes gradient magnitude — a "free" depth cue
- Depth-modulated line weight is the simplest enhancement to add to any contour plot
- Animation resolves all ambiguity but sacrifices static reproducibility
- Cross-hatching mimics classical engraving; line density patterns encode curvature direction
- Parametric iso-curves are the only viable approach for non-orientable surfaces
