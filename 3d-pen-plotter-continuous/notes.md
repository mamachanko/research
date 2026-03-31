# 3D Pen Plotter Continuous Stroke — Research Notes

## Goal
Explore how to draw three-dimensional objects using a pen plotter in a single continuous stroke without lifting the pen, while maintaining the perception of depth and three dimensions.

## Key Challenges
1. **Single continuous path**: A wireframe has many edges meeting at vertices — most wireframes are NOT Eulerian (not all vertices have even degree), so you can't simply trace all edges without retracing some.
2. **3D depth perception**: Without shading/fill, we rely on perspective projection, line weight variation, and hidden-line handling to convey depth.
3. **Artistic quality**: The path must look intentional, not like a random walk.

## Approach: Pipeline

### Step 1: Define 3D geometry
- Parametric shapes (cube, icosahedron, torus, sphere wireframe, etc.)
- Represented as vertices + edges (graph)

### Step 2: Project to 2D
- Perspective projection with a virtual camera
- Retain Z-depth for each vertex for depth cues

### Step 3: Eulerize the graph (Chinese Postman Problem)
- Find odd-degree vertices
- Compute shortest paths between all pairs of odd-degree vertices
- Find minimum-weight perfect matching on odd vertices
- Duplicate edges along matched shortest paths
- Now all vertices have even degree → Euler circuit exists

### Step 4: Find Euler circuit
- Hierholzer's algorithm — O(|E|) time

### Step 5: Render with depth cues
- Line weight varies with Z-depth (closer = thicker)
- Optional: opacity variation
- Output as SVG for pen plotter compatibility

## Mathematical Foundations

### Perspective Projection
For a point (x, y, z) with camera at origin looking down -z:
```
x_screen = f * (x / z) + cx
y_screen = f * (y / z) + cy
```

### Euler Circuit Conditions
- Undirected connected graph has Euler circuit ⟺ every vertex has even degree
- Euler path ⟺ exactly 0 or 2 vertices have odd degree

### Chinese Postman Problem
1. Find odd-degree vertices S
2. Shortest paths between all pairs in S
3. Minimum-weight perfect matching on S
4. Duplicate edges along matched paths
5. Hierholzer's algorithm on augmented graph

### Hierholzer's Algorithm
```
stack = [start], circuit = []
while stack:
    v = stack[-1]
    if v has unused edges:
        pick unused edge (v, u), mark used, stack.push(u)
    else:
        stack.pop(), circuit.append(v)
return reversed(circuit)
```

## Depth Cues for Single-Stroke Line Art
- **Line weight**: w(z) = w_max - (w_max - w_min) * (z - z_near) / (z_far - z_near)
- **Perspective foreshortening**: naturally occurs from projection
- **Back-face culling**: skip edges of back-facing-only faces to reduce clutter
- **Partial hidden line removal**: artistically, showing some hidden lines as thinner adds depth

## Artistic Considerations
- Retraced edges (from Eulerization) add visual weight, which can be artistically desirable for closer edges
- The starting point of the circuit affects the visual flow
- Space-filling curves mapped onto surfaces could replace wireframes for organic shapes
- Cross-hatching density as a function of surface normal angle to light creates shading in single-stroke

## Implementation Notes

### Matching Algorithm Performance
- Bitmask DP works for ≤22 odd-degree vertices (2^22 states fits in memory)
- Dodecahedron (20 odd-degree vertices) uses DP: ~1M states, runs in milliseconds
- Torus/Sphere/Möbius strip use greedy nearest-neighbor matching (too many odd vertices)
- Greedy is suboptimal but produces acceptable results for artistic purposes

### Results Summary
| Shape | Vertices | Edges | Retraced | Total Segments |
|-------|----------|-------|----------|----------------|
| Cube | 8 | 12 | 4 | 16 |
| Tetrahedron | 4 | 6 | 2 | 8 |
| Octahedron | 6 | 12 | 0 | 12 |
| Icosahedron | 12 | 30 | 6 | 36 |
| Dodecahedron | 20 | 30 | 10 | 40 |
| Torus | 160 | 320 | 0 | 320 |
| Sphere | 86 | 180 | 0 | 180 |
| Stella Octangula | 8 | 12 | 6 | 18 |
| Hexagonal Prism | 12 | 18 | 6 | 24 |
| Pentagonal Prism | 10 | 15 | 5 | 20 |
| Möbius Strip | 96 | 168 | 24 | 192 |

### Key Observations
- Graphs with all even-degree vertices (octahedron, torus, sphere) need zero retracing — they are naturally Eulerian
- The tetrahedron has 4 odd-degree vertices → needs 2 extra edges (one matching pair)
- The stella octangula (two disjoint tetrahedra) required bridge edges to connect the components before Eulerization
- Cube: 8 vertices all degree 3 → 4 extra edges. Dodecahedron: 20 vertices all degree 3 → 10 extra edges
- Perspective projection creates convincing foreshortening especially visible on the torus and sphere
- Depth-varying line weight (0.4px–3.5px range) with opacity (0.3–1.0) creates strong near/far perception
- The retraced edges add visual weight, which serendipitously reinforces certain structural lines
