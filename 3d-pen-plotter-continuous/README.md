# 3D Pen Plotter — Continuous Stroke

Drawing three-dimensional objects in a single, continuous stroke without lifting the pen, while maintaining the perception of depth through line weight and opacity variation.

## The Problem

A pen plotter draws by moving a pen across paper. The simplest approach is to draw each edge of a 3D wireframe as a separate line segment, lifting the pen between edges. But what if we want to draw the entire object in **one continuous stroke** — never lifting the pen?

This is a graph theory problem. A wireframe model is a graph where vertices are points in 3D space and edges are the lines connecting them. Drawing every edge exactly once without lifting the pen requires an **Euler path** (or **Euler circuit** if we return to the start). Such a path exists only when every vertex has even degree (for a circuit) or exactly two vertices have odd degree (for a path).

Most 3D wireframes don't satisfy this condition. A cube, for example, has 8 vertices each of degree 3 — all odd. The solution: strategically duplicate some edges to make all degrees even, then find the Euler circuit through the augmented graph. The pen retraces those duplicated edges, but the result is a single continuous stroke that covers every edge.

## Approach

### Pipeline

```
3D Geometry → Rotation → Perspective Projection → Component Bridging → Eulerization → Euler Circuit → SVG with Depth Cues
```

### Step 1: Define 3D Geometry

Eleven shapes are implemented as vertex + edge lists:

- **Platonic solids**: tetrahedron, cube, octahedron, icosahedron, dodecahedron
- **Other polyhedra**: stella octangula (two interlocking tetrahedra), pentagonal prism, hexagonal prism
- **Curved surfaces**: torus, sphere wireframe, Mobius strip

### Step 2: Perspective Projection

A virtual camera with configurable position, target, field of view, and up-vector transforms 3D coordinates to 2D screen space using perspective division. The Z-depth of each vertex in camera space is retained for depth-based rendering.

### Step 3: Connect Disconnected Components

Some shapes (like the stella octangula = two separate tetrahedra) consist of multiple disconnected subgraphs. Bridge edges are added between the nearest vertices of each component to create a single connected graph.

### Step 4: Eulerize the Graph (Chinese Postman Problem)

1. **Find odd-degree vertices** — vertices where an odd number of edges meet
2. **Compute shortest paths** between all pairs of odd-degree vertices (BFS)
3. **Find minimum-weight perfect matching** — pair up odd vertices to minimize total duplicated edge count
   - Bitmask DP for ≤22 odd vertices: O(n^2 * 2^n), exact optimal
   - Greedy nearest-neighbor matching for larger sets: fast but approximate
4. **Duplicate edges** along the matched shortest paths

After Eulerization, every vertex has even degree, guaranteeing an Euler circuit exists.

### Step 5: Find Euler Circuit (Hierholzer's Algorithm)

Hierholzer's algorithm finds the circuit in O(|E|) time by walking unused edges and splicing sub-tours.

### Step 6: Render with Depth Cues

Each segment of the continuous path is rendered with:
- **Line weight**: thicker for nearer edges (3.5px), thinner for far edges (0.4px)
- **Opacity**: stronger for near (1.0), fainter for far (0.3)
- **Perspective foreshortening**: naturally from the projection

A red dot marks the starting point of each continuous stroke.

## Results

| Shape | Vertices | Original Edges | Retraced | Total Segments |
|-------|----------|---------------|----------|----------------|
| Tetrahedron | 4 | 6 | 2 | 8 |
| Cube | 8 | 12 | 4 | 16 |
| Octahedron | 6 | 12 | 0 | 12 |
| Icosahedron | 12 | 30 | 6 | 36 |
| Dodecahedron | 20 | 30 | 10 | 40 |
| Torus | 160 | 320 | 0 | 320 |
| Sphere | 86 | 180 | 0 | 180 |
| Stella Octangula | 8 | 12 | 6 | 18 |
| Hexagonal Prism | 12 | 18 | 6 | 24 |
| Pentagonal Prism | 10 | 15 | 5 | 20 |
| Mobius Strip | 96 | 168 | 24 | 192 |

**Naturally Eulerian shapes** (zero retracing): the octahedron, torus, and sphere wireframe all have exclusively even-degree vertices — every vertex sits at the junction of an even number of edges — so an Euler circuit exists on the original graph with no edge duplication needed.

**Overhead from Eulerization**: for the Platonic solids where every vertex has degree 3 (tetrahedron, cube, icosahedron, dodecahedron), the number of retraced edges equals half the number of odd-degree vertices. This is the theoretical minimum.

## Key Findings

- **Depth perception from line weight alone is surprisingly effective.** A 9:1 ratio between the thickest and thinnest strokes (3.5px vs 0.4px), combined with opacity variation, creates a convincing sense of near vs. far even without hidden-line removal.

- **The Chinese Postman Problem is the right abstraction.** It directly minimizes the number of edges the pen must retrace. For wireframes of common 3D shapes, the number of odd-degree vertices is small enough (≤20) for exact optimal matching via bitmask DP.

- **Not all wireframes need Eulerization.** Regular grids (torus, sphere latitude/longitude mesh) naturally have all even-degree vertices because each interior vertex connects to exactly 4 neighbors. These shapes can be drawn in one stroke with zero wasted movement.

- **Disconnected components require bridge edges.** Compound shapes like the stella octangula (two interlocking tetrahedra that share no edges) need explicit connections added before Eulerization. The bridge edges become visible "travel lines" in the drawing — an artistic choice that traces a path through space between the two shapes.

- **Retraced edges aren't wasted — they add visual weight.** On a real pen plotter, retraced edges receive a double layer of ink, making them slightly bolder. This unintentional emphasis often falls on structurally important edges (connecting odd-degree vertices, which tend to be at corners or intersections).

## Files

- `plotter.py` — Complete implementation: geometry generation, projection, Eulerization, Euler circuit, SVG rendering (~800 lines, Python 3.8+ stdlib only)
- `gallery.html` — HTML gallery page embedding all SVGs
- `*.svg` — Individual SVG outputs for each shape
- `notes.md` — Working notes and observations from the investigation

## Running

```bash
cd 3d-pen-plotter-continuous
python3 plotter.py
# Generates 11 SVGs + gallery.html
```

No external dependencies required — uses only Python standard library.

## Pen Plotter Compatibility

The SVG output is designed for pen plotter workflows:
- All geometry is vector (line segments), no rasterization
- Coordinates are in screen pixels but easily scaled
- Compatible with tools like [vpype](https://github.com/abey79/vpype) for post-processing (optimizing, scaling, converting to HPGL/G-code)
- Line weight variation can be approximated on single-pen plotters by multi-pass: trace near (bold) edges 2-3 times, far edges once
