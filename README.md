# Research

Research projects carried out by AI tools.

Each directory here is a separate research project carried out by an LLM tool - usually Claude Code. Every single line of text and code was written by an LLM.

Times shown are in UTC.

### [3D Pen Plotter — Continuous Stroke](https://github.com/mamachanko/research/tree/main/3d-pen-plotter-continuous#readme) (2026-03-31 05:45)

An exploration of drawing three-dimensional wireframe objects in a single continuous pen stroke without lifting the pen, using the Chinese Postman Problem to Eulerize wireframe graphs and Hierholzer's algorithm for Euler circuits. Eleven shapes rendered as SVGs with depth-varying line weight and opacity for 3D perception.

Key findings:
- Depth perception from line weight variation alone (9:1 thick-to-thin ratio plus opacity) is surprisingly effective without hidden-line removal
- The Chinese Postman Problem (minimum-weight perfect matching on odd-degree vertices) is the right abstraction for minimizing retraced edges
- Regular grid wireframes (torus, sphere) are naturally Eulerian — zero retracing needed — because all vertices have even degree
- Retraced edges add visual weight at structurally important vertices, an unintentional but artistically pleasing side effect

### [Isoline 3D Visualization](https://github.com/mamachanko/research/tree/main/isoline-3d-visualization#readme) (2026-03-30 04:10)

An exploration of eight distinct approaches to visualizing 3D objects using only isolines: horizontal slicing, Joy Division ridgelines, rotating parametric curves, depth-layered contours, cross-hatched multi-axis contours, radial coordinate isolines, parametric surface grids, and animated morphing contours. Each technique is demonstrated on multiple mathematical surfaces with Python/matplotlib.

Key findings:
- Occlusion (Joy Division style) creates the strongest immediate 3D perception from pure line work
- Cross-hatching multiple contour sets from different axes mimics classical engraving and reveals curvature direction
- Depth-modulated line weight (thicker = closer) is the simplest enhancement to add depth to any contour plot
- Parametric iso-curves are the only viable approach for non-orientable surfaces like Klein bottles and Mobius strips

### [Visualize Queue Flow](https://github.com/mamachanko/research/tree/main/visualize-queue-flow#readme) (2026-03-29 09:20)

An exploration of 22 terminal visualizations for a two-app RabbitMQ pipeline (publisher → queue → consumer), built with Bubble Tea and Lip Gloss. Eight conceptual areas (queue depth, message flow, status panels, throughput, event timeline, network topology, latency distribution, full dashboard) were each implemented at three complexity levels, from a single colored progress bar to a particle burst stream to a 2D latency heat grid, all recorded as animated GIFs with VHS.

Key findings:
- Queue depth is the anchor metric — color-ramped fill bars communicate urgency faster than numbers, and every other panel becomes more legible when paired with it.
- Particle flow (messages as moving dots slowing into a queue zone) is the most immediately intuitive visualization of async buffering; even the static L1 version conveys the topology at a glance.
- 2D heat grids (latency bucket × time) reveal temporal patterns that snapshot histograms hide — rising queue pressure visibly shifts the hot-zone upward into higher-latency buckets.
- Sinusoidal pub/consume rates on different periods create natural fill/drain cycles with no manual scripting, keeping every recording visually dynamic.
- Playwright-bundled ffmpeg lacks the GIF muxer; Pillow's alpha-composite + palette-quantize pipeline is a clean drop-in replacement for VHS's native GIF output.

### [Spinner Design Variations](https://github.com/mamachanko/research/tree/main/spinner-design-variations#readme) (2026-03-29 05:52)

An exploration of animated terminal spinner designs inspired by the [crush CLI](https://github.com/charmbracelet/crush) gradient spinner. Eight variations were built in Go using HCL color blending via `go-colorful`, covering themes from Matrix Rain and Fire to Ocean Wave and Neon Glitch, each recorded as a GIF with VHS.

Key findings:
- HCL color blending produces perceptually smoother gradients than RGB, especially across hue boundaries.
- Character set density (braille, katakana vs. ASCII punctuation) strongly influences the perceived "weight" and mood of a spinner independent of color.
- Staggered birth delays (`BirthDelay`) have an outsized effect on personality — long delays feel organic, short ones feel snappy.
- Pre-rendering all frames upfront keeps the animation loop allocation-free and avoids per-frame color math.
