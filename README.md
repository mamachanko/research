# Research

Research projects carried out by AI tools.

Each directory here is a separate research project carried out by an LLM tool - usually Claude Code. Every single line of text and code was written by an LLM.

Times shown are in UTC.

### [AI Agent Optimization for Large Java Codebases](https://github.com/mamachanko/research/tree/main/ai-java-codebase-optimization#readme) (2026-04-03 12:00)

How to steer weaker AI models (e.g. Gemini 3.1 Pro) to produce top-quality outcomes on Java 17 / Maven multi-module / Spring Boot 4 codebases, using techniques applicable to any agent tooling (Cursor CLI, Claude Code, Codex, etc.). Covers context engineering, task decomposition, compile feedback loops, ArchUnit guardrails, and Spring Boot 4 specifics.

Key findings:
- A Project Bible file (AGENTS.md / .cursorrules / CLAUDE.md) injected at session start is the single highest-ROI lever — a weak model with precise context beats a strong model operating blind
- Module-scoped sessions (one Maven module per agent conversation) dramatically reduce failure radius and context noise
- Analysis-before-implementation prompting and contract-first interface definition eliminate the majority of hallucinated class names and incorrect signatures
- Machine-enforced guardrails (ArchUnit + Checkstyle) catch `javax.*` / cross-module / `@Transactional` violations more reliably than model self-policing
- Immediate compile-loop feedback (mvn compile errors fed back per edit) is the fastest path from broken to working code

### [Shinkawa Pen Plotter Style](https://github.com/mamachanko/research/tree/main/shinkawa-pen-plotter-style#readme) (2026-04-02 12:45)

An investigation into translating Yoji Shinkawa's ink brush aesthetic (Metal Gear Solid, Death Stranding) to pen plotter output. Analyzed his visual style — bold sumi-e-influenced contours, variable line weight, dry brush texture, ink splatter — and developed algorithmic strategies to approximate each element using plotter hardware. Includes a proof-of-concept Python script that generates multi-layer plotter-ready SVGs.

Key findings:
- Multi-layer decomposition (contours, flow strokes, dry brush, hatching, splatter) with per-layer pen swaps is the most practical approach
- Physical pen choice (actual Pentel brush pens in the plotter arm) matters more than algorithm sophistication
- Perlin noise flow fields seeded in dark image regions naturally produce gestural, Shinkawa-like stroke quality
- Plotter imperfections (ink flow variation, brush flex, paper texture) are features that add organic quality
- High-contrast "less is more" compositions (drawing only the darkest 20-30%) best capture Shinkawa's economy of mark-making

### [3D Pen Plotter — Continuous Stroke](https://github.com/mamachanko/research/tree/main/3d-pen-plotter-continuous#readme) (2026-03-31 11:20)

An exploration of drawing three-dimensional forms with a single continuous swirling stroke — not wireframes, but flowing parametric curves that wrap around 3D surfaces. When projected with perspective, the line's natural density variation (bunching at edges, spreading in the middle) makes the brain perceive spheres, tori, vases, and knots. Twelve shapes rendered as SVGs with depth-varying line weight and opacity.

Key findings:
- A single spiral on a sphere's surface, projected with perspective, reads unmistakably as a 3D sphere from line density alone
- Line weight variation (14:1 thick-to-thin ratio) combined with opacity is the strongest depth cue — no hidden-line removal needed
- Three different sphere spirals (pole-to-pole, loxodrome, Fibonacci) each produce distinct artistic character while all clearly reading as "sphere"
- The technique extends to any parametric surface: tori, cones, vases, Mobius strips, Klein bottles, and knots

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
